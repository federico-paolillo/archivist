namespace Archivist.Gateway.Application.Articles;

/// <summary>
/// Represents the outcome of an authenticated article hard-delete request.
/// </summary>
public enum ArticleDeleteResult
{
    /// <summary>The article, associated rows, and artifact directory were removed.</summary>
    Deleted,

    /// <summary>The article does not exist for the authenticated user.</summary>
    NotFound,

    /// <summary>The article has at least one running job and cannot be deleted.</summary>
    RunningJobConflict,

    /// <summary>The artifact directory could not be removed, so database state was left intact.</summary>
    ArtifactCleanupFailed,
}