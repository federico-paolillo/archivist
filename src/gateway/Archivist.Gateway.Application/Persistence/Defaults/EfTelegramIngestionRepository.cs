using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;

namespace Archivist.Gateway.Application.Persistence.Defaults;

/// <summary>
/// EF Core implementation of Telegram ingestion persistence.
/// </summary>
public sealed class EfTelegramIngestionRepository(
    ArchivistDbContext db,
    IUlidGenerator ids,
    TimeProvider timeProvider) : ITelegramIngestionRepository
{
    private const int SqliteConstraintErrorCode = 19;

    /// <inheritdoc />
    public async Task<RecordTelegramIngestionResult> RecordValidUrlAsync(
        RecordTelegramIngestionCommand command,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(command);
        ArgumentException.ThrowIfNullOrWhiteSpace(command.OriginalUrl);

        var existing = await db.Jobs.AsNoTracking()
            .Where(x => x.TelegramUpdateId == command.TelegramUpdateId)
            .Select(x => new { x.ArticleId, x.Id })
            .SingleOrDefaultAsync(cancellationToken)
            .ConfigureAwait(false);

        if (existing is not null)
        {
            return new RecordTelegramIngestionResult(false, existing.ArticleId, existing.Id);
        }

        await using var transaction = await db.Database.BeginTransactionAsync(cancellationToken).ConfigureAwait(false);

        await db.Database.ExecuteSqlInterpolatedAsync(
                $"""
                INSERT INTO users (id, telegram_user_id)
                VALUES ({PersistenceConstants.PersonalUserId}, {command.TelegramUserId})
                ON CONFLICT(id) DO UPDATE SET telegram_user_id = excluded.telegram_user_id;
                """,
                cancellationToken)
            .ConfigureAwait(false);
        db.ChangeTracker.Clear();

        var now = timeProvider.GetUtcNow();
        var article = new ArticleEntity
        {
            Id = ids.NewId(),
            UserId = PersistenceConstants.PersonalUserId,
            OriginalUrl = command.OriginalUrl,
            Status = PersistenceConstants.ArticleQueued,
            CreatedAt = now,
        };
        var job = new JobEntity
        {
            Id = ids.NewId(),
            UserId = PersistenceConstants.PersonalUserId,
            ArticleId = article.Id,
            Type = PersistenceConstants.ArticleProcessingJobType,
            Status = PersistenceConstants.JobQueued,
            TelegramUpdateId = command.TelegramUpdateId,
            TelegramChatId = command.TelegramChatId,
            TelegramMessageId = command.TelegramMessageId,
            TelegramUserId = command.TelegramUserId,
            CreatedAt = now,
        };

        db.Articles.Add(article);
        db.Jobs.Add(job);

        try
        {
            await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
        }
        catch (DbUpdateException ex) when (IsTelegramUpdateUniqueConstraintViolation(ex))
        {
            await transaction.RollbackAsync(CancellationToken.None).ConfigureAwait(false);
            db.ChangeTracker.Clear();

            var duplicate = await db.Jobs.AsNoTracking()
                .Where(x => x.TelegramUpdateId == command.TelegramUpdateId)
                .Select(x => new { x.ArticleId, x.Id })
                .SingleOrDefaultAsync(cancellationToken)
                .ConfigureAwait(false);

            if (duplicate is null)
            {
                throw;
            }

            return new RecordTelegramIngestionResult(false, duplicate.ArticleId, duplicate.Id);
        }

        await transaction.CommitAsync(cancellationToken).ConfigureAwait(false);

        return new RecordTelegramIngestionResult(true, article.Id, job.Id);
    }

    private static bool IsTelegramUpdateUniqueConstraintViolation(DbUpdateException exception)
    {
        return exception.GetBaseException() is SqliteException
        {
            SqliteErrorCode: SqliteConstraintErrorCode,
            Message: var message,
        } &&
            message.Contains("UNIQUE constraint failed: jobs.telegram_update_id", StringComparison.Ordinal);
    }
}