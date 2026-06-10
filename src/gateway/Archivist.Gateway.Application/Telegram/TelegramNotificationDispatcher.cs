namespace Archivist.Gateway.Application.Telegram;

using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.Observability;
using Archivist.Gateway.Application.Persistence;

using Microsoft.Extensions.Logging;

/// <summary>
/// Dispatches pending terminal Telegram notifications from SQLite notification rows.
/// For failed jobs: sends error_message as the reply body.
/// For succeeded jobs: reads summary.md and sends it as the reply body.
/// </summary>
public sealed partial class TelegramNotificationDispatcher(
    ITelegramNotificationRepository notificationRepository,
    IArticleArtifactReader artifactReader,
    ITelegramClient telegramClient,
    TimeProvider timeProvider,
    ILogger<TelegramNotificationDispatcher> logger)
{
    /// <summary>
    /// Telegram message length limit in characters.
    /// </summary>
    public const int TelegramMessageMaxLength = 4096;

    private static readonly TimeSpan NotificationTtl = TimeSpan.FromDays(7);
    private const string SummarySuccessPrefix = "Archived. Summary is:";
    private const string NotificationIdAttribute = "archivist.notification.id";

    /// <summary>
    /// Polls pending notification rows and dispatches terminal Telegram replies for terminal jobs.
    /// </summary>
    public async Task DispatchPendingAsync(CancellationToken cancellationToken)
    {
        var pending = await notificationRepository.GetPendingAsync(cancellationToken).ConfigureAwait(false);

        foreach (var notification in pending)
        {
            await ProcessOneAsync(notification, cancellationToken).ConfigureAwait(false);
        }
    }

    /// <summary>
    /// Deletes sent and failed notifications whose expiry has passed.
    /// </summary>
    public async Task CleanUpExpiredAsync(CancellationToken cancellationToken)
    {
        var now = timeProvider.GetUtcNow();
        await notificationRepository.DeleteExpiredAsync(now, cancellationToken).ConfigureAwait(false);
    }

    private async Task ProcessOneAsync(PendingNotificationRow notification, CancellationToken cancellationToken)
    {
        using var activity = ArchivistTelemetry.ActivitySource.StartActivity("gateway.telegram.notification_dispatch");
        activity?.SetTag(NotificationIdAttribute, notification.NotificationId);
        activity?.SetTag(ArchivistTelemetry.JobId, notification.JobId);
        activity?.SetTag(ArchivistTelemetry.ArticleId, notification.ArticleId);
        activity?.SetTag(ArchivistTelemetry.Stage, "telegram_notification_dispatch");

        if (notification.JobStatus == PersistenceConstants.JobSucceeded)
        {
            await DispatchSucceededAsync(notification, cancellationToken).ConfigureAwait(false);
            return;
        }

        if (notification.JobStatus != PersistenceConstants.JobFailed)
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "non_terminal");
            LogSkippedNonTerminal(logger, notification.NotificationId, notification.JobId, notification.JobStatus);
            return;
        }

        if (!await EnsureReplyTargetAsync(notification, cancellationToken).ConfigureAwait(false))
        {
            return;
        }

        var errorText = notification.JobErrorMessage ?? string.Empty;
        var replyText = Truncate(errorText);

        await SendReplyAndMarkAsync(notification, replyText, cancellationToken).ConfigureAwait(false);
    }

    private async Task DispatchSucceededAsync(PendingNotificationRow notification, CancellationToken cancellationToken)
    {
        if (!await EnsureReplyTargetAsync(notification, cancellationToken).ConfigureAwait(false))
        {
            return;
        }

        string summary;
        try
        {
            using var reader = await artifactReader
                .OpenSummaryMarkdownAsync(notification.ArticleId, cancellationToken)
                .ConfigureAwait(false);
            summary = await reader.ReadToEndAsync(cancellationToken).ConfigureAwait(false);
        }
        catch (Exception ex) when (ex is ArticleArtifactReadException or IOException or UnauthorizedAccessException)
        {
            const string artifactError = "Summary artifact missing or unreadable.";
            var failedAt = timeProvider.GetUtcNow();

            await notificationRepository
                .MarkFailedAsync(notification.NotificationId, artifactError, failedAt, failedAt.Add(NotificationTtl), cancellationToken)
                .ConfigureAwait(false);

            System.Diagnostics.Activity.Current?.SetTag(ArchivistTelemetry.Outcome, "summary_artifact_read_failed");
            System.Diagnostics.Activity.Current?.SetStatus(System.Diagnostics.ActivityStatusCode.Error, "summary artifact read failed");
            System.Diagnostics.Activity.Current?.AddException(ex);
            LogSummaryArtifactReadFailed(logger, ex, notification.NotificationId, notification.JobId, notification.ArticleId);
            return;
        }

        var replyText = Truncate($"{SummarySuccessPrefix} {summary}");

        await SendReplyAndMarkAsync(notification, replyText, cancellationToken).ConfigureAwait(false);
    }

    private async Task<bool> EnsureReplyTargetAsync(PendingNotificationRow notification, CancellationToken cancellationToken)
    {
        if (notification.TelegramChatId is not null && notification.TelegramMessageId is not null)
        {
            return true;
        }

        const string missingTargetError = "Missing Telegram reply target: telegram_chat_id or telegram_message_id is null.";
        var failedAt = timeProvider.GetUtcNow();

        await notificationRepository
            .MarkFailedAsync(notification.NotificationId, missingTargetError, failedAt, failedAt.Add(NotificationTtl), cancellationToken)
            .ConfigureAwait(false);

        System.Diagnostics.Activity.Current?.SetTag(ArchivistTelemetry.Outcome, "missing_reply_target");
        LogMissingReplyTarget(logger, notification.NotificationId, notification.JobId);
        return false;
    }

    private async Task SendReplyAndMarkAsync(PendingNotificationRow notification, string replyText, CancellationToken cancellationToken)
    {
        var chatId = notification.TelegramChatId!.Value;
        var messageId = notification.TelegramMessageId!.Value;
        var now = timeProvider.GetUtcNow();
        var expiresAt = now.Add(NotificationTtl);

        try
        {
            await telegramClient.SendReplyAsync(chatId, messageId, replyText, cancellationToken).ConfigureAwait(false);

            await notificationRepository.MarkSentAsync(notification.NotificationId, now, expiresAt, cancellationToken).ConfigureAwait(false);

            System.Diagnostics.Activity.Current?.SetTag(ArchivistTelemetry.Outcome, "sent");
            LogSent(logger, notification.NotificationId, notification.JobId);
        }
#pragma warning disable CA1031 // Telegram delivery failure must be recorded and not propagate; any exception is caught here per spec REQ-026.
        catch (Exception ex)
#pragma warning restore CA1031
        {
            var deliveryError = $"Telegram delivery failed: {ex.Message}";

            await notificationRepository
                .MarkFailedAsync(notification.NotificationId, deliveryError, now, expiresAt, cancellationToken)
                .ConfigureAwait(false);

            System.Diagnostics.Activity.Current?.SetTag(ArchivistTelemetry.Outcome, "delivery_failed");
            System.Diagnostics.Activity.Current?.SetStatus(System.Diagnostics.ActivityStatusCode.Error, "telegram delivery failed");
            System.Diagnostics.Activity.Current?.AddException(ex);
            LogDeliveryFailed(logger, ex, notification.NotificationId, notification.JobId);
        }
    }

    public static string Truncate(string text)
    {
        ArgumentNullException.ThrowIfNull(text);

        if (text.Length <= TelegramMessageMaxLength)
        {
            return text;
        }

        const string ellipsis = "…";
        return string.Concat(text.AsSpan(0, TelegramMessageMaxLength - ellipsis.Length), ellipsis);
    }

    [LoggerMessage(Level = LogLevel.Warning, Message = "Notification {NotificationId} for job {JobId}: skipping non-terminal job status {JobStatus}")]
    private static partial void LogSkippedNonTerminal(ILogger logger, string notificationId, string jobId, string jobStatus);

    [LoggerMessage(Level = LogLevel.Warning, Message = "Notification {NotificationId} for job {JobId}: missing Telegram reply target; skipping")]
    private static partial void LogMissingReplyTarget(ILogger logger, string notificationId, string jobId);

    [LoggerMessage(Level = LogLevel.Information, Message = "Notification {NotificationId} for job {JobId}: sent successfully")]
    private static partial void LogSent(ILogger logger, string notificationId, string jobId);

    [LoggerMessage(Level = LogLevel.Error, Message = "Notification {NotificationId} for job {JobId}: Telegram delivery failed; notification marked failed")]
    private static partial void LogDeliveryFailed(ILogger logger, Exception ex, string notificationId, string jobId);

    [LoggerMessage(Level = LogLevel.Error, Message = "Notification {NotificationId} for job {JobId}: summary artifact read failed for article {ArticleId}; notification marked failed")]
    private static partial void LogSummaryArtifactReadFailed(ILogger logger, Exception ex, string notificationId, string jobId, string articleId);
}