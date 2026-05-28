namespace Archivist.Gateway.Api.Articles.Models;

/// <summary>
/// Article detail response including UI-readable Markdown artifacts.
/// </summary>
internal sealed record ArticleDetailResponse(
    string Id,
    string? Title,
    string OriginalUrl,
    string? CanonicalUrl,
    string Status,
    string? ErrorMessage,
    DateTimeOffset CreatedAt,
    string? SummaryMarkdown,
    string? ContentMarkdown);