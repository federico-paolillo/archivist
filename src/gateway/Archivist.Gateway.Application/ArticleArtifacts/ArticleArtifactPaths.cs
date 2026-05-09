namespace Archivist.Gateway.Application.ArticleArtifacts;

/// <summary>
/// Resolves deterministic article artifact paths from DATA_DIR and article ids.
/// </summary>
public sealed class ArticleArtifactPaths(string dataDirectory)
{
    /// <summary>
    /// Gets the absolute article artifact directory.
    /// </summary>
    public string ArticleDirectory(string articleId) => Path.Combine(dataDirectory, "articles", ValidateArticleId(articleId));

    /// <summary>
    /// Gets the deterministic HTML snapshot path.
    /// </summary>
    public string SnapshotHtml(string articleId) => Path.Combine(ArticleDirectory(articleId), "snapshot.html");

    /// <summary>
    /// Gets the deterministic Markdown content path.
    /// </summary>
    public string ContentMarkdown(string articleId) => Path.Combine(ArticleDirectory(articleId), "content.md");

    /// <summary>
    /// Gets the deterministic summary Markdown path.
    /// </summary>
    public string SummaryMarkdown(string articleId) => Path.Combine(ArticleDirectory(articleId), "summary.md");

    /// <summary>
    /// Gets the deterministic future summary JSON path.
    /// </summary>
    public string SummaryJson(string articleId) => Path.Combine(ArticleDirectory(articleId), "summary.json");

    /// <summary>
    /// Gets the deterministic future metadata path.
    /// </summary>
    public string MetadataJson(string articleId) => Path.Combine(ArticleDirectory(articleId), "metadata.json");

    private static string ValidateArticleId(string articleId)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(articleId);

        if (articleId.Length != 26 || articleId.Any(c => !char.IsAsciiLetterOrDigit(c)))
        {
            throw new ArgumentException("Article id must be a ULID path segment.", nameof(articleId));
        }

        return articleId;
    }
}