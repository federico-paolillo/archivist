namespace Archivist.Gateway.Application.Persistence.Entities;

/// <summary>
/// Represents a worker processing job.
/// </summary>
public sealed class JobEntity
{
    /// <summary>
    /// Gets or sets the job ULID.
    /// </summary>
    public required string Id { get; set; }

    /// <summary>
    /// Gets or sets the owner user ULID.
    /// </summary>
    public required string UserId { get; set; }

    /// <summary>
    /// Gets or sets the article ULID.
    /// </summary>
    public required string ArticleId { get; set; }

    /// <summary>
    /// Gets or sets the job type.
    /// </summary>
    public required string Type { get; set; }

    /// <summary>
    /// Gets or sets queued, running, succeeded, or failed.
    /// </summary>
    public required string Status { get; set; }

    /// <summary>
    /// Gets or sets Telegram update idempotency metadata.
    /// </summary>
    public long? TelegramUpdateId { get; set; }

    /// <summary>
    /// Gets or sets Telegram reply chat metadata.
    /// </summary>
    public long? TelegramChatId { get; set; }

    /// <summary>
    /// Gets or sets Telegram reply message metadata.
    /// </summary>
    public long? TelegramMessageId { get; set; }

    /// <summary>
    /// Gets or sets Telegram sender identity metadata.
    /// </summary>
    public long? TelegramUserId { get; set; }

    /// <summary>
    /// Gets or sets the public terminal error message.
    /// </summary>
    public string? ErrorMessage { get; set; }

    /// <summary>
    /// Gets or sets the creation timestamp.
    /// </summary>
    public DateTimeOffset CreatedAt { get; set; }

    /// <summary>
    /// Gets or sets the worker start timestamp.
    /// </summary>
    public DateTimeOffset? StartedAt { get; set; }

    /// <summary>
    /// Gets or sets the terminal timestamp.
    /// </summary>
    public DateTimeOffset? CompletedAt { get; set; }

    /// <summary>
    /// Gets or sets the terminal cleanup eligibility timestamp.
    /// </summary>
    public DateTimeOffset? ExpiresAt { get; set; }
}