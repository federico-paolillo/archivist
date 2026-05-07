namespace Archivist.Gateway.Application.Persistence.Entities;

using System.Diagnostics.CodeAnalysis;

/// <summary>
/// Represents durable article state.
/// </summary>
public sealed class ArticleEntity
{
    /// <summary>
    /// Gets or sets the article ULID.
    /// </summary>
    public required string Id { get; set; }

    /// <summary>
    /// Gets or sets the owner user ULID.
    /// </summary>
    public required string UserId { get; set; }

    /// <summary>
    /// Gets or sets the submitted URL.
    /// </summary>
    [SuppressMessage("Design", "CA1056:URI-like properties should not be strings", Justification = "SQLite persistence stores URL text columns.")]
    public required string OriginalUrl { get; set; }

    /// <summary>
    /// Gets or sets the canonical URL discovered by processing.
    /// </summary>
    [SuppressMessage("Design", "CA1056:URI-like properties should not be strings", Justification = "SQLite persistence stores URL text columns.")]
    public string? CanonicalUrl { get; set; }

    /// <summary>
    /// Gets or sets the article title discovered by processing.
    /// </summary>
    public string? Title { get; set; }

    /// <summary>
    /// Gets or sets queued, ready, or failed.
    /// </summary>
    public required string Status { get; set; }

    /// <summary>
    /// Gets or sets the public terminal error message.
    /// </summary>
    public string? ErrorMessage { get; set; }

    /// <summary>
    /// Gets or sets the creation timestamp.
    /// </summary>
    public DateTimeOffset CreatedAt { get; set; }
}