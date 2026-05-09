using Archivist.Gateway.Application.Persistence;

using Microsoft.EntityFrameworkCore;

using Xunit.Abstractions;

namespace Archivist.Gateway.Tests.Persistence;

public sealed class GatewayPersistenceCompositionTest(ITestOutputHelper testOutputHelper) : IntegrationTest(testOutputHelper)
{
    [Fact]
    public async Task GatewayCreatesSchemaWhenSQLitePathIsConfigured()
    {
        var sqlitePath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");

        PrepareEnvironment(
            Environments.Development,
            configureConfiguration: configuration =>
                configuration.AddInMemoryCollection(new Dictionary<string, string?>
                {
                    ["SQLITE_PATH"] = sqlitePath,
                }));

        using var client = CreateHttpClient();

        using var db = new ArchivistDbContext(
            new DbContextOptionsBuilder<ArchivistDbContext>()
                .UseSqlite($"Data Source={sqlitePath}")
                .Options);

        Assert.False(await db.Users.AnyAsync(CancellationToken.None));
    }
}