namespace Archivist.Gateway.Application.Articles;

/// <summary>
/// Loads authenticated user-owned article metadata and detail.
/// </summary>
public interface IArticleQueryService
{
    /// <summary>
    /// Lists a fixed-size page of article metadata for the specified owner.
    /// </summary>
    Task<ArticleListPage> ListAsync(
        string userId,
        string? after,
        string? before,
        CancellationToken cancellationToken);

    /// <summary>
    /// Loads article detail and Markdown artifacts for the specified owner.
    /// </summary>
    Task<ArticleDetailResult> GetDetailAsync(string articleId, string userId, CancellationToken cancellationToken);
}