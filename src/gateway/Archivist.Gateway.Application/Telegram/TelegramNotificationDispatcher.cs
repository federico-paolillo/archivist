namespace Archivist.Gateway.Application.Telegram;

using Archivist.Gateway.Application.Persistence;

using Microsoft.Extensions.Logging;

/// <summary>
/// Dispatches pending terminal Telegram notifications from SQLite notification rows.
/// For failed jobs: sends error_message as the reply body.
/// For succeeded jobs: leaves the notification pending until a downstream feature such as SUMGEN-005 provides success content.
/// </summary>
public sealed partial class TelegramNotificationDispatcher(
    ITelegramNotificationRepository notificationRepository,
    ITelegramClient telegramClient,
    TimeProvider timeProvider,
    ILogger<TelegramNotificationDispatcher> logger)
{
    /// <summary>
    /// Telegram message length limit in characters.
    /// </summary>
    public const int TelegramMessageMaxLength = 4096;

    private static readonly TimeSpan NotificationTtl = TimeSpan.FromDays(7);

    /// <summary>
    /// Polls pending notification rows and dispatches terminal Telegram replies for failed jobs.
    /// Succeeded-job notifications remain pending until SUMGEN-005 provides success content.
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
        if (notification.JobStatus == PersistenceConstants.JobSucceeded)
        {
            LogSucceededJobDeferred(logger, notification.NotificationId, notification.JobId);
            return;
        }

        if (notification.JobStatus != PersistenceConstants.JobFailed)
        {
            LogSkippedNonTerminal(logger, notification.NotificationId, notification.JobId, notification.JobStatus);
            return;
        }

        if (notification.TelegramChatId is null || notification.TelegramMessageId is null)
        {
            const string missingTargetError = "Missing Telegram reply target: telegram_chat_id or telegram_message_id is null.";
            var failedAt = timeProvider.GetUtcNow();

            await notificationRepository
                .MarkFailedAsync(notification.NotificationId, missingTargetError, failedAt, failedAt.Add(NotificationTtl), cancellationToken)
                .ConfigureAwait(false);

            LogMissingReplyTarget(logger, notification.NotificationId, notification.JobId);
            return;
        }

        var chatId = notification.TelegramChatId.Value;
        var messageId = notification.TelegramMessageId.Value;
        var errorText = notification.JobErrorMessage ?? string.Empty;
        var replyText = Truncate(errorText);

        var now = timeProvider.GetUtcNow();
        var expiresAt = now.Add(NotificationTtl);

        try
        {
            await telegramClient.SendReplyAsync(chatId, messageId, replyText, cancellationToken).ConfigureAwait(false);

            await notificationRepository.MarkSentAsync(notification.NotificationId, now, expiresAt, cancellationToken).ConfigureAwait(false);

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

    [LoggerMessage(Level = LogLevel.Debug, Message = "Notification {NotificationId} for succeeded job {JobId}: deferred until SUMGEN-005 provides success content")]
    private static partial void LogSucceededJobDeferred(ILogger logger, string notificationId, string jobId);

    [LoggerMessage(Level = LogLevel.Warning, Message = "Notification {NotificationId} for job {JobId}: skipping non-terminal job status {JobStatus}")]
    private static partial void LogSkippedNonTerminal(ILogger logger, string notificationId, string jobId, string jobStatus);

    [LoggerMessage(Level = LogLevel.Warning, Message = "Notification {NotificationId} for job {JobId}: missing Telegram reply target; skipping")]
    private static partial void LogMissingReplyTarget(ILogger logger, string notificationId, string jobId);

    [LoggerMessage(Level = LogLevel.Information, Message = "Notification {NotificationId} for job {JobId}: sent successfully")]
    private static partial void LogSent(ILogger logger, string notificationId, string jobId);

    [LoggerMessage(Level = LogLevel.Error, Message = "Notification {NotificationId} for job {JobId}: Telegram delivery failed; notification marked failed")]
    private static partial void LogDeliveryFailed(ILogger logger, Exception ex, string notificationId, string jobId);
}