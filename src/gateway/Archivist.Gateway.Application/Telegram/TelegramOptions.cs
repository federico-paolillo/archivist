namespace Archivist.Gateway.Application.Telegram;

/// <summary>
/// Configuration options for Telegram integration.
/// </summary>
public sealed class TelegramOptions
{
    /// <summary>
    /// Gets or sets the Telegram bot token used for API calls.
    /// </summary>
    public string BotToken { get; set; } = string.Empty;

    /// <summary>
    /// Gets or sets the webhook secret used to validate incoming updates.
    /// </summary>
    public string WebhookSecret { get; set; } = string.Empty;

    /// <summary>
    /// Gets or sets the allowed Telegram user ID for ingestion.
    /// </summary>
    public long AllowedUserId { get; set; }
}