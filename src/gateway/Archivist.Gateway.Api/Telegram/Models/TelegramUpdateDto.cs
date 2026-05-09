namespace Archivist.Gateway.Api.Telegram.Models;

using System.Text.Json.Serialization;

/// <summary>
/// Represents a Telegram Bot API update delivered by webhook.
/// </summary>
public sealed class TelegramUpdateDto
{
    /// <summary>
    /// Gets or sets the Telegram update identifier.
    /// </summary>
    [JsonPropertyName("update_id")]
    public long UpdateId { get; set; }

    /// <summary>
    /// Gets or sets the message, if present.
    /// </summary>
    [JsonPropertyName("message")]
    public TelegramMessageDto? Message { get; set; }
}

/// <summary>
/// Represents a Telegram message.
/// </summary>
public sealed class TelegramMessageDto
{
    /// <summary>
    /// Gets or sets the message identifier.
    /// </summary>
    [JsonPropertyName("message_id")]
    public long MessageId { get; set; }

    /// <summary>
    /// Gets or sets the chat that the message was sent in.
    /// </summary>
    [JsonPropertyName("chat")]
    public TelegramChatDto? Chat { get; set; }

    /// <summary>
    /// Gets or sets the sender user.
    /// </summary>
    [JsonPropertyName("from")]
    public TelegramUserDto? From { get; set; }

    /// <summary>
    /// Gets or sets the message text.
    /// </summary>
    [JsonPropertyName("text")]
    public string? Text { get; set; }
}

/// <summary>
/// Represents a Telegram chat.
/// </summary>
public sealed class TelegramChatDto
{
    /// <summary>
    /// Gets or sets the chat identifier.
    /// </summary>
    [JsonPropertyName("id")]
    public long Id { get; set; }
}

/// <summary>
/// Represents a Telegram user.
/// </summary>
public sealed class TelegramUserDto
{
    /// <summary>
    /// Gets or sets the Telegram user identifier.
    /// </summary>
    [JsonPropertyName("id")]
    public long Id { get; set; }
}