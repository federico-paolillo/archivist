using System.Data;

using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.Observability;
using Archivist.Gateway.Application.Persistence;

using Microsoft.EntityFrameworkCore;

namespace Archivist.Gateway.Application.Articles.Defaults;

/// <summary>
/// EF Core implementation of authenticated article hard deletion.
/// </summary>
public sealed class EfArticleDeleteService(
    ArchivistDbContext db,
    IArticleArtifactDeletion artifactDeletion,
    TimeProvider timeProvider) : IArticleDeleteService
{
    private static readonly TimeSpan ForceDeleteStaleThreshold = TimeSpan.FromHours(2);

    /// <inheritdoc />
    public async Task<ArticleDeleteResult> DeleteAsync(
        string articleId,
        string userId,
        CancellationToken cancellationToken)
    {
        return await DeleteCoreAsync(
                articleId,
                userId,
                allowStaleRunningJobs: false,
                cancellationToken)
            .ConfigureAwait(false);
    }

    /// <inheritdoc />
    public async Task<ArticleDeleteResult> ForceDeleteAsync(
        string articleId,
        string userId,
        CancellationToken cancellationToken)
    {
        return await DeleteCoreAsync(
                articleId,
                userId,
                allowStaleRunningJobs: true,
                cancellationToken)
            .ConfigureAwait(false);
    }

    private async Task<ArticleDeleteResult> DeleteCoreAsync(
        string articleId,
        string userId,
        bool allowStaleRunningJobs,
        CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(articleId);
        ArgumentException.ThrowIfNullOrWhiteSpace(userId);

        using var activity = ArchivistTelemetry.ActivitySource.StartActivity("gateway.articles.delete");
        activity?.SetTag(ArchivistTelemetry.ArticleId, articleId);
        activity?.SetTag(ArchivistTelemetry.Stage, allowStaleRunningJobs ? "articles_force_delete" : "articles_delete");

        await using var transaction = await db.Database
            .BeginTransactionAsync(IsolationLevel.Serializable, cancellationToken)
            .ConfigureAwait(false);

        try
        {
            var articleExists = await db.Articles
                .AsNoTracking()
                .AnyAsync(x => x.Id == articleId && x.UserId == userId, cancellationToken)
                .ConfigureAwait(false);

            if (!articleExists)
            {
                await transaction.RollbackAsync(cancellationToken).ConfigureAwait(false);
                activity?.SetTag(ArchivistTelemetry.Outcome, "not_found");

                return ArticleDeleteResult.NotFound;
            }

            var runningJobStartTimes = await db.Jobs
                .AsNoTracking()
                .Where(x =>
                    x.ArticleId == articleId &&
                    x.UserId == userId &&
                    x.Status == PersistenceConstants.JobRunning)
                .Select(x => x.StartedAt)
                .ToListAsync(cancellationToken)
                .ConfigureAwait(false);

            if (runningJobStartTimes.Count > 0 &&
                (!allowStaleRunningJobs || runningJobStartTimes.Any(IsActiveRunningJob)))
            {
                await transaction.RollbackAsync(cancellationToken).ConfigureAwait(false);
                activity?.SetTag(ArchivistTelemetry.Outcome, "running_job_conflict");

                return ArticleDeleteResult.RunningJobConflict;
            }

            var jobIds = await db.Jobs
                .Where(x => x.ArticleId == articleId && x.UserId == userId)
                .Select(x => x.Id)
                .ToListAsync(cancellationToken)
                .ConfigureAwait(false);

            await db.Notifications
                .Where(x => jobIds.Contains(x.JobId))
                .ExecuteDeleteAsync(cancellationToken)
                .ConfigureAwait(false);

            await db.Jobs
                .Where(x => x.ArticleId == articleId && x.UserId == userId)
                .ExecuteDeleteAsync(cancellationToken)
                .ConfigureAwait(false);

            await db.Articles
                .Where(x => x.Id == articleId && x.UserId == userId)
                .ExecuteDeleteAsync(cancellationToken)
                .ConfigureAwait(false);

            var artifactDeleted = await artifactDeletion
                .DeleteArticleDirectoryAsync(articleId, cancellationToken)
                .ConfigureAwait(false);

            if (!artifactDeleted)
            {
                await transaction.RollbackAsync(cancellationToken).ConfigureAwait(false);
                activity?.SetTag(ArchivistTelemetry.Outcome, "artifact_cleanup_failed");

                return ArticleDeleteResult.ArtifactCleanupFailed;
            }

            await transaction.CommitAsync(cancellationToken).ConfigureAwait(false);
            activity?.SetTag(ArchivistTelemetry.Outcome, "deleted");

            return ArticleDeleteResult.Deleted;
        }
        catch
        {
            await transaction.RollbackAsync(CancellationToken.None).ConfigureAwait(false);
            activity?.SetStatus(System.Diagnostics.ActivityStatusCode.Error);
            throw;
        }
    }

    private bool IsActiveRunningJob(DateTimeOffset? startedAt)
    {
        if (startedAt is null)
        {
            return false;
        }

        return startedAt > timeProvider.GetUtcNow().Subtract(ForceDeleteStaleThreshold);
    }
}