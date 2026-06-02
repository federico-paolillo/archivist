namespace Archivist.Gateway.Application.Articles;

/// <summary>
/// Deletes authenticated user-owned article state and deterministic artifacts.
/// </summary>
public interface IArticleDeleteService
{
    /// <summary>
    /// Deletes the article and associated state for the specified owner.
    /// </summary>
    Task<ArticleDeleteResult> DeleteAsync(string articleId, string userId, CancellationToken cancellationToken);

    /// <summary>
    /// Deletes the article and associated state for the specified owner when running jobs are stale.
    /// </summary>
    Task<ArticleDeleteResult> ForceDeleteAsync(string articleId, string userId, CancellationToken cancellationToken);
}