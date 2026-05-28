namespace Archivist.Gateway.Application.ArticleArtifacts;

/// <summary>
/// Reads deterministic article artifacts owned by the Worker.
/// </summary>
public interface IArticleArtifactReader
{
    /// <summary>
    /// Reads the final v0 summary Markdown artifact for an article.
    /// </summary>
    Task<string> ReadSummaryMarkdownAsync(string articleId, CancellationToken cancellationToken);
}