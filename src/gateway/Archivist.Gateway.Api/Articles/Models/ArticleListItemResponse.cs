namespace Archivist.Gateway.Api.Articles.Models;

/// <summary>
/// Article metadata item returned by the list endpoint.
/// </summary>
internal sealed record ArticleListItemResponse(
    string Id,
    string? Title,
    string OriginalUrl,
    string? CanonicalUrl,
    string Status,
    string? ErrorMessage,
    DateTimeOffset CreatedAt);