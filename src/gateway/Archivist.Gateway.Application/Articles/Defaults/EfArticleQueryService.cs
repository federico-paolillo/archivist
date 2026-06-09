using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.Observability;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.EntityFrameworkCore;

namespace Archivist.Gateway.Application.Articles.Defaults;

/// <summary>
/// EF Core implementation of authenticated article read queries.
/// </summary>
public sealed class EfArticleQueryService(
    ArchivistDbContext db,
    IArticleArtifactReader artifactReader,
    TimeProvider timeProvider) : IArticleQueryService
{
    private const int PageSize = 25;
    private static readonly TimeSpan ForceDeleteStaleThreshold = TimeSpan.FromHours(2);

    /// <inheritdoc />
    public async Task<ArticleListPage> ListAsync(
        string userId,
        string? after,
        string? before,
        CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(userId);

        using var activity = ArchivistTelemetry.ActivitySource.StartActivity("gateway.articles.list");
        activity?.SetTag(ArchivistTelemetry.UserId, userId);
        activity?.SetTag(ArchivistTelemetry.Stage, "articles_list");

        var query = db.Articles
            .AsNoTracking()
            .Where(x => x.UserId == userId);

        if (after is not null)
        {
            query = query.Where(x => x.Id.CompareTo(after) < 0);
        }

        if (before is not null)
        {
            query = query.Where(x => x.Id.CompareTo(before) > 0);
        }

        var items = await query
            .OrderByDescending(x => x.Id)
            .Take(PageSize)
            .Select(x => new ArticleListItem(
                x.Id,
                x.Title,
                x.OriginalUrl,
                x.CanonicalUrl,
                x.Status,
                x.ErrorMessage,
                x.CreatedAt))
            .ToListAsync(cancellationToken)
            .ConfigureAwait(false);

        var nextCursor = await GetNextCursorAsync(userId, items, cancellationToken).ConfigureAwait(false);
        var previousCursor = await GetPreviousCursorAsync(userId, items, cancellationToken).ConfigureAwait(false);

        activity?.SetTag(ArchivistTelemetry.Outcome, "found");

        return new ArticleListPage(items, nextCursor, previousCursor);
    }

    /// <inheritdoc />
    public async Task<ArticleDetailResult> GetDetailAsync(
        string articleId,
        string userId,
        CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(articleId);
        ArgumentException.ThrowIfNullOrWhiteSpace(userId);

        using var activity = ArchivistTelemetry.ActivitySource.StartActivity("gateway.articles.detail");
        activity?.SetTag(ArchivistTelemetry.ArticleId, articleId);
        activity?.SetTag(ArchivistTelemetry.UserId, userId);
        activity?.SetTag(ArchivistTelemetry.Stage, "articles_detail");

        var article = await db.Articles
            .AsNoTracking()
            .Where(x => x.Id == articleId && x.UserId == userId)
            .SingleOrDefaultAsync(cancellationToken)
            .ConfigureAwait(false);

        if (article is null)
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "not_found");

            return ArticleDetailResult.NotFound;
        }

        var summaryMarkdown = await ReadArtifactAsync(
                article,
                artifactReader.OpenSummaryMarkdownAsync,
                cancellationToken)
            .ConfigureAwait(false);

        var contentMarkdown = await ReadArtifactAsync(
                article,
                artifactReader.OpenContentMarkdownAsync,
                cancellationToken)
            .ConfigureAwait(false);

        if (article.Status == PersistenceConstants.ArticleReady &&
            (summaryMarkdown is null || contentMarkdown is null))
        {
            activity?.SetTag(ArchivistTelemetry.Outcome, "artifact_unavailable");

            return ArticleDetailResult.RequiredArtifactUnavailable;
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
        var canForceDelete = runningJobStartTimes.Count > 0 &&
            runningJobStartTimes.All(IsStaleRunningJob);

        activity?.SetTag(ArchivistTelemetry.Outcome, "found");

        return ArticleDetailResult.Found(new ArticleDetail(
            article.Id,
            article.Title,
            article.OriginalUrl,
            article.CanonicalUrl,
            article.Status,
            article.ErrorMessage,
            article.CreatedAt,
            summaryMarkdown,
            contentMarkdown,
            canForceDelete));
    }

    private bool IsStaleRunningJob(DateTimeOffset? startedAt)
    {
        if (startedAt is null)
        {
            return true;
        }

        return startedAt <= timeProvider.GetUtcNow().Subtract(ForceDeleteStaleThreshold);
    }

    private async Task<string?> ReadArtifactAsync(
        ArticleEntity article,
        Func<string, CancellationToken, Task<TextReader>> open,
        CancellationToken cancellationToken)
    {
        try
        {
            using var reader = await open(article.Id, cancellationToken).ConfigureAwait(false);
            return await reader.ReadToEndAsync(cancellationToken).ConfigureAwait(false);
        }
        catch (Exception ex) when (ex is ArticleArtifactReadException or IOException or UnauthorizedAccessException)
        {
            return null;
        }
    }

    private async Task<string?> GetNextCursorAsync(
        string userId,
        List<ArticleListItem> items,
        CancellationToken cancellationToken)
    {
        if (items.Count == 0)
        {
            return null;
        }

        var lastId = items[^1].Id;
        var hasOlder = await db.Articles
            .AsNoTracking()
            .AnyAsync(x => x.UserId == userId && x.Id.CompareTo(lastId) < 0, cancellationToken)
            .ConfigureAwait(false);

        return hasOlder ? lastId : null;
    }

    private async Task<string?> GetPreviousCursorAsync(
        string userId,
        List<ArticleListItem> items,
        CancellationToken cancellationToken)
    {
        if (items.Count == 0)
        {
            return null;
        }

        var firstId = items[0].Id;
        var hasNewer = await db.Articles
            .AsNoTracking()
            .AnyAsync(x => x.UserId == userId && x.Id.CompareTo(firstId) > 0, cancellationToken)
            .ConfigureAwait(false);

        return hasNewer ? firstId : null;
    }
}