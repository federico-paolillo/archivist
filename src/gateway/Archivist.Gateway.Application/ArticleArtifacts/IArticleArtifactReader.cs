namespace Archivist.Gateway.Application.ArticleArtifacts;

/// <summary>
/// Reads deterministic article artifacts owned by the Worker.
/// </summary>
public interface IArticleArtifactReader
{
    /// <summary>
    /// Opens the extracted Markdown content artifact for an article.
    /// </summary>
    Task<TextReader> OpenContentMarkdownAsync(string articleId, CancellationToken cancellationToken);

    /// <summary>
    /// Opens the final v0 summary Markdown artifact for an article.
    /// </summary>
    Task<TextReader> OpenSummaryMarkdownAsync(string articleId, CancellationToken cancellationToken);
}