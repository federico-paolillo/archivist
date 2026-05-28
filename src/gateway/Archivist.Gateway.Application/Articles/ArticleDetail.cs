using System.Diagnostics.CodeAnalysis;

namespace Archivist.Gateway.Application.Articles;

/// <summary>
/// Represents one article detail row plus UI-readable Markdown artifacts.
/// </summary>
[SuppressMessage("Design", "CA1054:URI-like parameters should not be strings", Justification = "The canonical API contract exposes persisted URL text.")]
[SuppressMessage("Design", "CA1056:URI-like properties should not be strings", Justification = "The canonical API contract exposes persisted URL text.")]
public sealed record ArticleDetail(
    string Id,
    string? Title,
    string OriginalUrl,
    string? CanonicalUrl,
    string Status,
    string? ErrorMessage,
    DateTimeOffset CreatedAt,
    string? SummaryMarkdown,
    string? ContentMarkdown);