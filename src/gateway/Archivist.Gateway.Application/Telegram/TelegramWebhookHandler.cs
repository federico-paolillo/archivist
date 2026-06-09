namespace Archivist.Gateway.Application.Telegram;

using System.Diagnostics;
using System.Diagnostics.CodeAnalysis;

using Archivist.Gateway.Application.Observability;
using Archivist.Gateway.Application.Persistence;

using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

using OpenTelemetry;
using OpenTelemetry.Context.Propagation;

/// <summary>
/// Processes incoming Telegram webhook updates: validates the secret, authorizes the sender,
/// validates URL-only text, persists valid URLs atomically, and sends immediate Telegram replies.
/// </summary>
public sealed partial class TelegramWebhookHandler(
    IOptions<TelegramSettings> settings,
    ITelegramUserResolver userResolver,
    ITelegramIngestionRepository ingestionRepository,
    ITelegramClient telegramClient,
    ILogger<TelegramWebhookHandler> logger)
{
    private const string AcknowledgementReply = "Ok, I will have a look";
    private const string InvalidUrlReply = "Nope, you must send only an URL";
    private static readonly TraceContextPropagator TraceContextPropagator = new();

    /// <summary>
    /// Processes a single Telegram webhook update.
    /// </summary>
    public async Task<TelegramWebhookResult> HandleAsync(
        TelegramWebhookCommand command,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(command);

        using var activity = ArchivistTelemetry.ActivitySource.StartActivity("gateway.telegram.webhook");
        activity?.SetTag(ArchivistTelemetry.TelegramUpdateId, command.UpdateId);
        activity?.SetTag(ArchivistTelemetry.Stage, "telegram_webhook");

        if (!IsSecretValid(command.WebhookSecret, settings.Value.WebhookSecret))
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "bad_secret");
            LogBadSecret(logger, command.UpdateId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.BadSecret);
        }

        if (command.SenderUserId is null)
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "unauthorized");
            LogUnauthorized(logger, command.UpdateId, command.SenderUserId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.Unauthorized);
        }

        var senderUserId = command.SenderUserId.Value;
        var userId = await userResolver.ResolveUserIdAsync(senderUserId, cancellationToken).ConfigureAwait(false);
        if (userId is null)
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "unauthorized");
            LogUnauthorized(logger, command.UpdateId, command.SenderUserId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.Unauthorized);
        }

        activity?.SetTag(ArchivistTelemetry.UserId, userId);
        using var userScope = logger.BeginScope(new Dictionary<string, object?>
        {
            [ArchivistTelemetry.UserId] = userId,
        });

        if (command.ChatId is null || command.MessageId is null)
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "no_message");
            LogNoMessage(logger, command.UpdateId, senderUserId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.NoMessage);
        }

        var chatId = command.ChatId.Value;
        var messageId = command.MessageId.Value;

        if (command.MessageText is null || !TryParseUrl(command.MessageText, out var url))
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "invalid_url");
            LogInvalidUrl(logger, command.UpdateId, senderUserId);

            await SendReplyFireAndForgetAsync(chatId, messageId, InvalidUrlReply, cancellationToken).ConfigureAwait(false);

            return new TelegramWebhookResult(TelegramWebhookOutcome.InvalidUrl);
        }

        LogValidUrl(logger, command.UpdateId, senderUserId, chatId, messageId, url);
        var traceCarrier = CaptureTraceCarrier();

        var ingestionCommand = new RecordTelegramIngestionCommand(
            TelegramUpdateId: command.UpdateId,
            TelegramChatId: chatId,
            TelegramMessageId: messageId,
            TelegramUserId: senderUserId,
            UserId: userId,
            OriginalUrl: url,
            TraceParent: traceCarrier.TraceParent,
            TraceState: traceCarrier.TraceState);

        var result = await ingestionRepository.RecordValidUrlAsync(ingestionCommand, cancellationToken).ConfigureAwait(false);
        activity?.SetTag(ArchivistTelemetry.ArticleId, result.ArticleId);
        activity?.SetTag(ArchivistTelemetry.JobId, result.JobId);

        if (!result.Created)
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "duplicate");
            LogDuplicate(logger, command.UpdateId, result.ArticleId, result.JobId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.Duplicate);
        }

        activity?.SetTag(ArchivistTelemetry.Outcome, "queued");
        LogEnqueued(logger, command.UpdateId, result.ArticleId, result.JobId);

        try
        {
            await telegramClient.SendReplyAsync(chatId, messageId, AcknowledgementReply, cancellationToken).ConfigureAwait(false);
        }
#pragma warning disable CA1031 // Acknowledgement failure must not roll back the persisted job; any exception is intentionally swallowed here.
        catch (Exception ex)
#pragma warning restore CA1031
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "queued_reply_failed");
            activity?.SetStatus(ActivityStatusCode.Error, "telegram acknowledgement failed");
            activity?.AddException(ex);
            LogAcknowledgementFailed(logger, ex, command.UpdateId, result.ArticleId, result.JobId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.QueuedReplyFailed);
        }

        return new TelegramWebhookResult(TelegramWebhookOutcome.Queued);
    }

    private static TraceCarrier CaptureTraceCarrier()
    {
        var carrier = new Dictionary<string, string>(StringComparer.OrdinalIgnoreCase);
        TraceContextPropagator.Inject(
            new PropagationContext(Activity.Current?.Context ?? default, Baggage.Current),
            carrier,
            static (target, key, value) => target[key] = value);

        carrier.TryGetValue("traceparent", out var traceParent);
        carrier.TryGetValue("tracestate", out var traceState);

        return new TraceCarrier(traceParent, traceState);
    }

    private sealed record TraceCarrier(string? TraceParent, string? TraceState);

    [SuppressMessage("Design", "CA1031:Do not catch general exception types", Justification = "Invalid-message reply failure must not propagate; exception is logged.")]
    private async Task SendReplyFireAndForgetAsync(
        long chatId,
        long messageId,
        string text,
        CancellationToken cancellationToken)
    {
        try
        {
            await telegramClient.SendReplyAsync(chatId, messageId, text, cancellationToken).ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            LogReplyFailed(logger, ex, chatId, messageId, text);
        }
    }

    private static bool IsSecretValid(string? provided, string configured)
    {
        if (string.IsNullOrEmpty(configured))
        {
            return false;
        }

        return string.Equals(provided, configured, StringComparison.Ordinal);
    }

    [SuppressMessage("Design", "CA1054:URI-like parameters should not be strings", Justification = "Returns canonical URL string for persistence.")]
    [SuppressMessage("Design", "CA1055:URI-like return values should not be strings", Justification = "Returns canonical URL string for persistence.")]
    private static bool TryParseUrl(string text, [NotNullWhen(true)] out string? url)
    {
        var trimmed = text.Trim();

        if (trimmed.Any(char.IsWhiteSpace))
        {
            url = null;
            return false;
        }

        if (!Uri.TryCreate(trimmed, UriKind.Absolute, out var uri))
        {
            url = null;
            return false;
        }

        if (uri.Scheme != Uri.UriSchemeHttp && uri.Scheme != Uri.UriSchemeHttps)
        {
            url = null;
            return false;
        }

        url = trimmed;
        return true;
    }

    [LoggerMessage(Level = LogLevel.Warning, Message = "Telegram webhook update {UpdateId} rejected: invalid secret")]
    private static partial void LogBadSecret(ILogger logger, long updateId);

    [LoggerMessage(Level = LogLevel.Warning, Message = "Telegram webhook update {UpdateId} rejected: sender {SenderId} is not mapped to an Archivist user")]
    private static partial void LogUnauthorized(ILogger logger, long updateId, long? senderId);

    [LoggerMessage(Level = LogLevel.Information, Message = "Telegram webhook update {UpdateId} from user {UserId}: no processable text message")]
    private static partial void LogNoMessage(ILogger logger, long updateId, long userId);

    [LoggerMessage(Level = LogLevel.Information, Message = "Telegram webhook update {UpdateId} from user {UserId}: message is not a valid URL")]
    private static partial void LogInvalidUrl(ILogger logger, long updateId, long userId);

    [LoggerMessage(Level = LogLevel.Information, Message = "Telegram webhook update {UpdateId} from user {UserId} chat {ChatId} message {MessageId}: valid URL {Url}")]
    private static partial void LogValidUrl(ILogger logger, long updateId, long userId, long chatId, long messageId, string url);

    [LoggerMessage(Level = LogLevel.Information, Message = "Telegram webhook update {UpdateId}: duplicate, article {ArticleId} job {JobId} already exist")]
    private static partial void LogDuplicate(ILogger logger, long updateId, string articleId, string jobId);

    [LoggerMessage(Level = LogLevel.Information, Message = "Telegram webhook update {UpdateId}: enqueued article {ArticleId} job {JobId}")]
    private static partial void LogEnqueued(ILogger logger, long updateId, string articleId, string jobId);

    [LoggerMessage(Level = LogLevel.Error, Message = "Telegram webhook update {UpdateId}: acknowledgement reply failed for article {ArticleId} job {JobId}; job persists")]
    private static partial void LogAcknowledgementFailed(ILogger logger, Exception ex, long updateId, string articleId, string jobId);

    [LoggerMessage(Level = LogLevel.Error, Message = "Telegram reply to chat {ChatId} message {MessageId} failed: {Text}")]
    private static partial void LogReplyFailed(ILogger logger, Exception ex, long chatId, long messageId, string text);
}