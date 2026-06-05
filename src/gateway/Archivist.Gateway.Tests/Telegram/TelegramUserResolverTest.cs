namespace Archivist.Gateway.Tests.Telegram;

using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Entities;
using Archivist.Gateway.Application.Telegram.Defaults;

using Microsoft.EntityFrameworkCore;

public sealed class TelegramUserResolverTest
{
    [Fact]
    public async Task ResolveUserIdAsync_MappedSender_ReturnsUserId()
    {
        await using var db = await CreateDbAsync();
        db.Users.Add(new UserEntity { Id = "01ASB2XFCZJY7WHZ2FNRTMQJCT", TelegramUserId = 99999 });
        await db.SaveChangesAsync();
        var resolver = new EfTelegramUserResolver(db);

        var userId = await resolver.ResolveUserIdAsync(99999, CancellationToken.None);

        Assert.Equal("01ASB2XFCZJY7WHZ2FNRTMQJCT", userId);
    }

    [Fact]
    public async Task ResolveUserIdAsync_UnmappedSender_ReturnsNull()
    {
        await using var db = await CreateDbAsync();
        db.Users.Add(new UserEntity { Id = "01ASB2XFCZJY7WHZ2FNRTMQJCT", TelegramUserId = 99999 });
        await db.SaveChangesAsync();
        var resolver = new EfTelegramUserResolver(db);

        var userId = await resolver.ResolveUserIdAsync(11111, CancellationToken.None);

        Assert.Null(userId);
    }

    private static async Task<ArchivistDbContext> CreateDbAsync()
    {
        var options = new DbContextOptionsBuilder<ArchivistDbContext>()
            .UseSqlite($"Data Source={Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db")}")
            .Options;
        var db = new ArchivistDbContext(options);
        await db.Database.EnsureCreatedAsync(CancellationToken.None);

        return db;
    }
}