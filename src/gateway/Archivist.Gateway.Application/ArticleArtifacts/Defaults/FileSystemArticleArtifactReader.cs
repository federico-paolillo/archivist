namespace Archivist.Gateway.Application.ArticleArtifacts.Defaults;

/// <summary>
/// Filesystem-backed read-only access to deterministic article artifacts.
/// </summary>
public sealed class FileSystemArticleArtifactReader(ArticleArtifactPaths paths) : IArticleArtifactReader
{
    /// <inheritdoc />
    public async Task<string> ReadContentMarkdownAsync(string articleId, CancellationToken cancellationToken)
    {
        try
        {
            return await File
                .ReadAllTextAsync(paths.ContentMarkdown(articleId), cancellationToken)
                .ConfigureAwait(false);
        }
        catch (Exception ex) when (ex is IOException or UnauthorizedAccessException or ArgumentException)
        {
            throw new ArticleArtifactReadException(
                $"Content artifact for article {articleId} is missing or unreadable.",
                ex);
        }
    }

    /// <inheritdoc />
    public async Task<string> ReadSummaryMarkdownAsync(string articleId, CancellationToken cancellationToken)
    {
        try
        {
            return await File
                .ReadAllTextAsync(paths.SummaryMarkdown(articleId), cancellationToken)
                .ConfigureAwait(false);
        }
        catch (Exception ex) when (ex is IOException or UnauthorizedAccessException or ArgumentException)
        {
            throw new ArticleArtifactReadException(
                $"Summary artifact for article {articleId} is missing or unreadable.",
                ex);
        }
    }
}