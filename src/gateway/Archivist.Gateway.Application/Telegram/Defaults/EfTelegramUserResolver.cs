namespace Archivist.Gateway.Application.Telegram.Defaults;

using Archivist.Gateway.Application.Persistence;

using Microsoft.EntityFrameworkCore;

/// <summary>
/// Resolves Telegram sender identities from the users table.
/// </summary>
public sealed class EfTelegramUserResolver(ArchivistDbContext db) : ITelegramUserResolver
{
    /// <inheritdoc />
    public async Task<string?> ResolveUserIdAsync(long telegramUserId, CancellationToken cancellationToken)
    {
        return await db.Users
            .AsNoTracking()
            .Where(x => x.TelegramUserId == telegramUserId)
            .Select(x => x.Id)
            .SingleOrDefaultAsync(cancellationToken)
            .ConfigureAwait(false);
    }
}