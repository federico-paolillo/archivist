namespace Archivist.Gateway.Application.ArticleArtifacts;

/// <summary>
/// Deletes deterministic article artifact directories for validated article ids.
/// </summary>
public interface IArticleArtifactDeletion
{
    /// <summary>
    /// Deletes the deterministic artifact directory for the article, treating a missing directory as success.
    /// </summary>
    Task<bool> DeleteArticleDirectoryAsync(string articleId, CancellationToken cancellationToken);
}