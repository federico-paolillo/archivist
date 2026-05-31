namespace Archivist.Gateway.Application.Telegram;

using System.Diagnostics.CodeAnalysis;

using Archivist.Gateway.Application.Persistence;

using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

/// <summary>
/// Processes incoming Telegram webhook updates: validates the secret, authorizes the sender,
/// validates URL-only text, persists valid URLs atomically, and sends immediate Telegram replies.
/// </summary>
public sealed partial class TelegramWebhookHandler(
    IOptions<TelegramSettings> settings,
    ITelegramIngestionRepository ingestionRepository,
    ITelegramClient telegramClient,
    ILogger<TelegramWebhookHandler> logger)
{
    private const string AcknowledgementReply = "Ok, I will have a look";
    private const string InvalidUrlReply = "Nope, you must send only an URL";

    /// <summary>
    /// Processes a single Telegram webhook update.
    /// </summary>
    public async Task<TelegramWebhookResult> HandleAsync(
        TelegramWebhookCommand command,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(command);

        var telegramSettings = settings.Value;

        if (!IsSecretValid(command.WebhookSecret, telegramSettings.WebhookSecret))
        {
            LogBadSecret(logger, command.UpdateId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.BadSecret);
        }

        if (command.SenderUserId is null || command.SenderUserId != telegramSettings.AllowedUserId)
        {
            LogUnauthorized(logger, command.UpdateId, command.SenderUserId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.Unauthorized);
        }

        if (command.ChatId is null || command.MessageId is null)
        {
            LogNoMessage(logger, command.UpdateId, command.SenderUserId.Value);

            return new TelegramWebhookResult(TelegramWebhookOutcome.NoMessage);
        }

        var chatId = command.ChatId.Value;
        var messageId = command.MessageId.Value;
        var senderUserId = command.SenderUserId.Value;

        if (command.MessageText is null || !TryParseUrl(command.MessageText, out var url))
        {
            LogInvalidUrl(logger, command.UpdateId, senderUserId);

            await SendReplyFireAndForgetAsync(chatId, messageId, InvalidUrlReply, cancellationToken).ConfigureAwait(false);

            return new TelegramWebhookResult(TelegramWebhookOutcome.InvalidUrl);
        }

        LogValidUrl(logger, command.UpdateId, senderUserId, chatId, messageId, url);

        var ingestionCommand = new RecordTelegramIngestionCommand(
            TelegramUpdateId: command.UpdateId,
            TelegramChatId: chatId,
            TelegramMessageId: messageId,
            TelegramUserId: senderUserId,
            OriginalUrl: url);

        var result = await ingestionRepository.RecordValidUrlAsync(ingestionCommand, cancellationToken).ConfigureAwait(false);

        if (!result.Created)
        {
            LogDuplicate(logger, command.UpdateId, result.ArticleId, result.JobId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.Duplicate);
        }

        LogEnqueued(logger, command.UpdateId, result.ArticleId, result.JobId);

        try
        {
            await telegramClient.SendReplyAsync(chatId, messageId, AcknowledgementReply, cancellationToken).ConfigureAwait(false);
        }
#pragma warning disable CA1031 // Acknowledgement failure must not roll back the persisted job; any exception is intentionally swallowed here.
        catch (Exception ex)
#pragma warning restore CA1031
        {
            LogAcknowledgementFailed(logger, ex, command.UpdateId, result.ArticleId, result.JobId);

            return new TelegramWebhookResult(TelegramWebhookOutcome.QueuedReplyFailed);
        }

        return new TelegramWebhookResult(TelegramWebhookOutcome.Queued);
    }

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

        if (trimmed.Contains(' ') || trimmed.Contains('\n') || trimmed.Contains('\r'))
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

    [LoggerMessage(Level = LogLevel.Warning, Message = "Telegram webhook update {UpdateId} rejected: sender {SenderId} is not the allowed user")]
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