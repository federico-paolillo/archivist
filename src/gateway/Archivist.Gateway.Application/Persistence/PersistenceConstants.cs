namespace Archivist.Gateway.Application.Persistence;

/// <summary>
/// Defines canonical persistence constants shared by gateway persistence contracts.
/// </summary>
public static class PersistenceConstants
{
    /// <summary>
    /// Article status used before worker processing completes.
    /// </summary>
    public const string ArticleQueued = "queued";

    /// <summary>
    /// Article status used after terminal success.
    /// </summary>
    public const string ArticleReady = "ready";

    /// <summary>
    /// Article status used after terminal failure.
    /// </summary>
    public const string ArticleFailed = "failed";

    /// <summary>
    /// Processing job type for article work.
    /// </summary>
    public const string ArticleProcessingJobType = "article-processing";

    /// <summary>
    /// Job status used before worker claim.
    /// </summary>
    public const string JobQueued = "queued";

    /// <summary>
    /// Job status used while worker processing owns the job.
    /// </summary>
    public const string JobRunning = "running";

    /// <summary>
    /// Job status used after terminal success.
    /// </summary>
    public const string JobSucceeded = "succeeded";

    /// <summary>
    /// Job status used after terminal failure.
    /// </summary>
    public const string JobFailed = "failed";

    /// <summary>
    /// Notification status used before gateway dispatch.
    /// </summary>
    public const string NotificationPending = "pending";

    /// <summary>
    /// Notification status used after successful dispatch.
    /// </summary>
    public const string NotificationSent = "sent";

    /// <summary>
    /// Notification status used after failed dispatch.
    /// </summary>
    public const string NotificationFailed = "failed";
}