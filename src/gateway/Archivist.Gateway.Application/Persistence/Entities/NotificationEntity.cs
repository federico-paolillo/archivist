namespace Archivist.Gateway.Application.Persistence.Entities;

/// <summary>
/// Represents a gateway terminal notification delivery record.
/// </summary>
public sealed class NotificationEntity
{
    /// <summary>
    /// Gets or sets the notification ULID.
    /// </summary>
    public required string Id { get; set; }

    /// <summary>
    /// Gets or sets the originating job ULID.
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
    /// Gets or sets the successful dispatch timestamp.
    /// </summary>
    public DateTimeOffset? SentAt { get; set; }

    /// <summary>
    /// Gets or sets the cleanup eligibility timestamp.
    /// </summary>
    public DateTimeOffset ExpiresAt { get; set; }
}