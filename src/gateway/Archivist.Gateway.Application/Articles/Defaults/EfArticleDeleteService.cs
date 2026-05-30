using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.Persistence;

using Microsoft.EntityFrameworkCore;

namespace Archivist.Gateway.Application.Articles.Defaults;

/// <summary>
/// EF Core implementation of authenticated article hard deletion.
/// </summary>
public sealed class EfArticleDeleteService(
    ArchivistDbContext db,
    IArticleArtifactDeletion artifactDeletion) : IArticleDeleteService
{
    /// <inheritdoc />
    public async Task<ArticleDeleteResult> DeleteAsync(
        string articleId,
        string userId,
        CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(articleId);
        ArgumentException.ThrowIfNullOrWhiteSpace(userId);

        await db.Database.OpenConnectionAsync(cancellationToken).ConfigureAwait(false);
        await db.Database.ExecuteSqlRawAsync("BEGIN IMMEDIATE;", cancellationToken).ConfigureAwait(false);

        try
        {
            var articleExists = await db.Articles
                .AsNoTracking()
                .AnyAsync(x => x.Id == articleId && x.UserId == userId, cancellationToken)
                .ConfigureAwait(false);

            if (!articleExists)
            {
                await db.Database.ExecuteSqlRawAsync("COMMIT;", cancellationToken).ConfigureAwait(false);
                return ArticleDeleteResult.NotFound;
            }

            var hasRunningJob = await db.Jobs
                .AsNoTracking()
                .AnyAsync(x =>
                    x.ArticleId == articleId &&
                    x.UserId == userId &&
                    x.Status == PersistenceConstants.JobRunning,
                    cancellationToken)
                .ConfigureAwait(false);

            if (hasRunningJob)
            {
                await db.Database.ExecuteSqlRawAsync("COMMIT;", cancellationToken).ConfigureAwait(false);
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
                await db.Database.ExecuteSqlRawAsync("ROLLBACK;", cancellationToken).ConfigureAwait(false);
                return ArticleDeleteResult.ArtifactCleanupFailed;
            }

            await db.Database.ExecuteSqlRawAsync("COMMIT;", cancellationToken).ConfigureAwait(false);
            return ArticleDeleteResult.Deleted;
        }
        catch
        {
            await db.Database.ExecuteSqlRawAsync("ROLLBACK;", CancellationToken.None).ConfigureAwait(false);
            throw;
        }
    }
}