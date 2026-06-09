namespace Archivist.Gateway.Application.ArticleArtifacts.Defaults;

/// <summary>
/// Filesystem-backed read-only access to deterministic article artifacts.
/// </summary>
public sealed class FileSystemArticleArtifactReader(ArticleArtifactPaths paths) : IArticleArtifactReader
{
    /// <inheritdoc />
    public Task<TextReader> OpenContentMarkdownAsync(string articleId, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();

        try
        {
            return Task.FromResult<TextReader>(OpenReader(paths.ContentMarkdown(articleId)));
        }
        catch (Exception ex) when (ex is IOException or UnauthorizedAccessException or ArgumentException)
        {
            throw new ArticleArtifactReadException(
                $"Content artifact for article {articleId} is missing or unreadable.",
                ex);
        }
    }

    /// <inheritdoc />
    public Task<TextReader> OpenSummaryMarkdownAsync(string articleId, CancellationToken cancellationToken)
    {
        cancellationToken.ThrowIfCancellationRequested();

        try
        {
            return Task.FromResult<TextReader>(OpenReader(paths.SummaryMarkdown(articleId)));
        }
        catch (Exception ex) when (ex is IOException or UnauthorizedAccessException or ArgumentException)
        {
            throw new ArticleArtifactReadException(
                $"Summary artifact for article {articleId} is missing or unreadable.",
                ex);
        }
    }

    private static StreamReader OpenReader(string path)
    {
        var stream = new FileStream(
            path,
            new FileStreamOptions
            {
                Mode = FileMode.Open,
                Access = FileAccess.Read,
                Share = FileShare.ReadWrite | FileShare.Delete,
                Options = FileOptions.Asynchronous | FileOptions.SequentialScan,
            });

        try
        {
            return new StreamReader(stream);
        }
        catch
        {
            stream.Dispose();
            throw;
        }
    }
}