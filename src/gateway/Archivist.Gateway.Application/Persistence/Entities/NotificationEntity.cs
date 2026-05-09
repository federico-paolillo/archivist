namespace Archivist.Gateway.Application.Persistence.Entities;

/// <summary>
/// Represents gateway terminal Telegram delivery state.
/// </summary>
public sealed class NotificationEntity
{
    /// <summary>
    /// Gets or sets the notification ULID.
    /// </summary>
    public required string Id { get; set; }

    /// <summary>
    /// Gets or sets the terminal job ULID.
    /// </summary>
    public required string JobId { get; set; }

    /// <summary>
    /// Gets or sets pending, sent, or failed.
    /// </summary>
    public required string Status { get; set; }

    /// <summary>
    /// Gets or sets the delivery error message.
    /// </summary>
    public string? ErrorMessage { get; set; }

    /// <summary>
    /// Gets or sets the creation timestamp.
    /// </summary>
    public DateTimeOffset CreatedAt { get; set; }

    /// <summary>
    /// Gets or sets the sent timestamp.
    /// </summary>
    public DateTimeOffset? SentAt { get; set; }

    /// <summary>
    /// Gets or sets the sent or failed cleanup eligibility timestamp.
    /// </summary>
    public DateTimeOffset ExpiresAt { get; set; }
}