namespace Archivist.Gateway.Application.Persistence.Defaults;

using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.EntityFrameworkCore;

/// <summary>
/// EF Core implementation of Telegram notification delivery persistence.
/// </summary>
public sealed class EfTelegramNotificationRepository(ArchivistDbContext db) : ITelegramNotificationRepository
{
    /// <inheritdoc />
    public async Task<IReadOnlyList<PendingNotificationRow>> GetPendingAsync(CancellationToken cancellationToken)
    {
        var rows = await db.Notifications
            .AsNoTracking()
            .Where(n => n.Status == PersistenceConstants.NotificationPending)
            .Join(
                db.Jobs.AsNoTracking(),
                n => n.JobId,
                j => j.Id,
                (n, j) => new PendingNotificationRow(
                    n.Id,
                    j.Id,
                    j.Status,
                    j.ErrorMessage,
                    j.TelegramChatId,
                    j.TelegramMessageId))
            .ToListAsync(cancellationToken)
            .ConfigureAwait(false);

        return rows;
    }

    /// <inheritdoc />
    public async Task MarkSentAsync(
        string notificationId,
        DateTimeOffset sentAt,
        DateTimeOffset expiresAt,
        CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(notificationId);

        var notification = await FindRequiredAsync(notificationId, cancellationToken).ConfigureAwait(false);
        notification.Status = PersistenceConstants.NotificationSent;
        notification.SentAt = sentAt;
        notification.ExpiresAt = expiresAt;

        await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
    }

    /// <inheritdoc />
    public async Task MarkFailedAsync(
        string notificationId,
        string errorMessage,
        DateTimeOffset failedAt,
        DateTimeOffset expiresAt,
        CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(notificationId);
        ArgumentException.ThrowIfNullOrWhiteSpace(errorMessage);

        var notification = await FindRequiredAsync(notificationId, cancellationToken).ConfigureAwait(false);
        notification.Status = PersistenceConstants.NotificationFailed;
        notification.ErrorMessage = errorMessage;
        notification.ExpiresAt = expiresAt;

        await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
    }

    /// <inheritdoc />
    public async Task DeleteExpiredAsync(DateTimeOffset now, CancellationToken cancellationToken)
    {
        // Load all non-pending notifications and filter by expiry on the client.
        // SQLite stores DateTimeOffset as text; server-side date comparison requires
        // format-aware SQLite functions not directly available through EF Core LINQ.
        var terminal = await db.Notifications
            .Where(n => n.Status != PersistenceConstants.NotificationPending)
            .ToListAsync(cancellationToken)
            .ConfigureAwait(false);

        var expired = terminal.Where(n => n.ExpiresAt <= now).ToList();

        if (expired.Count > 0)
        {
            db.Notifications.RemoveRange(expired);
            await db.SaveChangesAsync(cancellationToken).ConfigureAwait(false);
        }
    }

    private async Task<NotificationEntity> FindRequiredAsync(string notificationId, CancellationToken cancellationToken)
    {
        var notification = await db.Notifications
            .SingleOrDefaultAsync(n => n.Id == notificationId, cancellationToken)
            .ConfigureAwait(false);

        if (notification is null)
        {
            throw new InvalidOperationException($"Notification {notificationId} not found.");
        }

        return notification;
    }
}