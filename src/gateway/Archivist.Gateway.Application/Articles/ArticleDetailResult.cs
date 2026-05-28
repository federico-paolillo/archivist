namespace Archivist.Gateway.Application.Articles;

/// <summary>
/// Represents the outcome of loading article detail for an authenticated user.
/// </summary>
public sealed record ArticleDetailResult(ArticleDetailResultKind Kind, ArticleDetail? Article)
{
    /// <summary>
    /// Creates a successful detail result.
    /// </summary>
    public static ArticleDetailResult Found(ArticleDetail article) => new(ArticleDetailResultKind.Found, article);

    /// <summary>
    /// Creates a not-found detail result.
    /// </summary>
    public static ArticleDetailResult NotFound { get; } = new(ArticleDetailResultKind.NotFound, null);

    /// <summary>
    /// Creates a result for a ready article whose required artifacts cannot be read.
    /// </summary>
    public static ArticleDetailResult RequiredArtifactUnavailable { get; } = new(ArticleDetailResultKind.RequiredArtifactUnavailable, null);
}