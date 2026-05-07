using Archivist.Gateway.Application.Persistence.Entities;

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
    /// <inheritdoc />
    public async Task<RecordTelegramIngestionResult> RecordValidUrlAsync(
        RecordTelegramIngestionCommand command,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(command);
        ArgumentException.ThrowIfNullOrWhiteSpace(command.OriginalUrl);

        await using var transaction = await db.Database.BeginTransactionAsync(cancellationToken).ConfigureAwait(false);

        var existing = await db.Jobs.AsNoTracking()
            .Where(x => x.TelegramUpdateId == command.TelegramUpdateId)
            .Select(x => new { x.ArticleId, x.Id })
            .SingleOrDefaultAsync(cancellationToken)
            .ConfigureAwait(false);

        if (existing is not null)
        {
            await transaction.CommitAsync(cancellationToken).ConfigureAwait(false);

            return new RecordTelegramIngestionResult(false, existing.ArticleId, existing.Id);
        }

        var user = await db.Users
            .SingleOrDefaultAsync(x => x.Id == PersistenceConstants.PersonalUserId, cancellationToken)
            .ConfigureAwait(false);

        if (user is null)
        {
            user = new UserEntity
            {
                Id = PersistenceConstants.PersonalUserId,
                TelegramUserId = command.TelegramUserId,
            };
            db.Users.Add(user);
        }
        else
        {
            user.TelegramUserId = command.TelegramUserId;
        }

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

        await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
        await transaction.CommitAsync(cancellationToken).ConfigureAwait(false);

        return new RecordTelegramIngestionResult(true, article.Id, job.Id);
    }
}