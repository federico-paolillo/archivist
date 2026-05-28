namespace Archivist.Gateway.Api.Articles.Models;

/// <summary>
/// Fixed-size article metadata page returned by the list endpoint.
/// </summary>
internal sealed record ArticleListResponse(
    IReadOnlyList<ArticleListItemResponse> Items,
    string? NextCursor,
    string? PreviousCursor);