namespace Archivist.Gateway.Application.Articles;

/// <summary>
/// Enumerates article detail load outcomes.
/// </summary>
public enum ArticleDetailResultKind
{
    /// <summary>The article exists for the authenticated user and detail was loaded.</summary>
    Found,

    /// <summary>The article does not exist for the authenticated user.</summary>
    NotFound,

    /// <summary>A ready article is missing a required readable artifact.</summary>
    RequiredArtifactUnavailable,
}