namespace Archivist.Gateway.Application.Telegram;

/// <summary>
/// Captures the data extracted from a Telegram webhook update for processing.
/// </summary>
public sealed record TelegramWebhookCommand(
    string? WebhookSecret,
    long UpdateId,
    long? SenderUserId,
    long? ChatId,
    long? MessageId,
    string? MessageText);