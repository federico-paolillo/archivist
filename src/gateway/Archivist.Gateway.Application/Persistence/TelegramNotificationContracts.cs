namespace Archivist.Gateway.Application.Persistence;

/// <summary>
/// Represents a pending notification joined with job Telegram reply metadata.
/// </summary>
public sealed record PendingNotificationRow(
    string NotificationId,
    string JobId,
    string JobStatus,
    string? JobErrorMessage,
    long? TelegramChatId,
    long? TelegramMessageId);

/// <summary>
/// Reads pending notifications and finalises their delivery state in SQLite.
/// </summary>
public interface ITelegramNotificationRepository
{
    /// <summary>
    /// Returns all pending notification rows joined with job Telegram metadata.
    /// </summary>
    Task<IReadOnlyList<PendingNotificationRow>> GetPendingAsync(CancellationToken cancellationToken);

    /// <summary>
    /// Marks a notification as sent and sets its expiry.
    /// </summary>
    Task MarkSentAsync(string notificationId, DateTimeOffset sentAt, DateTimeOffset expiresAt, CancellationToken cancellationToken);

    /// <summary>
    /// Marks a notification as failed with a delivery error and sets its expiry.
    /// </summary>
    Task MarkFailedAsync(string notificationId, string errorMessage, DateTimeOffset failedAt, DateTimeOffset expiresAt, CancellationToken cancellationToken);

    /// <summary>
    /// Deletes sent and failed notifications whose expiry is in the past.
    /// </summary>
    Task DeleteExpiredAsync(DateTimeOffset now, CancellationToken cancellationToken);
}