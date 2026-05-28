namespace Archivist.Gateway.Application.Articles;

/// <summary>
/// Represents a fixed-size article metadata page and navigation cursors.
/// </summary>
public sealed record ArticleListPage(
    IReadOnlyList<ArticleListItem> Items,
    string? NextCursor,
    string? PreviousCursor);