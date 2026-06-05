namespace Archivist.Gateway.Application.Telegram;

/// <summary>
/// Resolves Telegram sender identities to persisted Archivist users.
/// </summary>
public interface ITelegramUserResolver
{
    /// <summary>
    /// Returns the Archivist user id mapped to the Telegram sender id, or <c>null</c> when no mapping exists.
    /// </summary>
    Task<string?> ResolveUserIdAsync(long telegramUserId, CancellationToken cancellationToken);
}