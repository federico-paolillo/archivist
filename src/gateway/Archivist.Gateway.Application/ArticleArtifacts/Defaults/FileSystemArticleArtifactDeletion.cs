namespace Archivist.Gateway.Application.ArticleArtifacts.Defaults;

/// <summary>
/// Filesystem-backed article artifact directory deletion.
/// </summary>
public sealed class FileSystemArticleArtifactDeletion(ArticleArtifactPaths paths) : IArticleArtifactDeletion
{
    /// <inheritdoc />
    public Task<bool> DeleteArticleDirectoryAsync(string articleId, CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(articleId);

        cancellationToken.ThrowIfCancellationRequested();

        try
        {
            var directory = paths.ArticleDirectory(articleId);

            if (!Directory.Exists(directory))
            {
                return Task.FromResult(true);
            }

            Directory.Delete(directory, recursive: true);
            return Task.FromResult(true);
        }
        catch (IOException)
        {
            return Task.FromResult(false);
        }
        catch (UnauthorizedAccessException)
        {
            return Task.FromResult(false);
        }
    }
}