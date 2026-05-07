using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Defaults;
using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Time.Testing;

namespace Archivist.Gateway.Tests.Persistence;

public sealed class TelegramIngestionRepositoryTest
{
    [Fact]
    public async Task RecordValidUrlCreatesUserArticleAndQueuedJobAtomically()
    {
        await using var db = await CreateDbAsync();
        db.Users.Add(new UserEntity
        {
            Id = PersistenceConstants.PersonalUserId,
            PasswordHash = "$argon2id$existing",
        });
        await db.SaveChangesAsync();

        var repo = CreateRepository(db);

        var result = await repo.RecordValidUrlAsync(
            new RecordTelegramIngestionCommand(
                100,
                200,
                300,
                400,
                "https://example.com/article"),
            CancellationToken.None);

        Assert.True(result.Created);

        var user = await db.Users.SingleAsync(CancellationToken.None);
        var article = await db.Articles.SingleAsync(CancellationToken.None);
        var job = await db.Jobs.SingleAsync(CancellationToken.None);

        Assert.Equal(PersistenceConstants.PersonalUserId, user.Id);
        Assert.Equal(400, user.TelegramUserId);
        Assert.Equal("$argon2id$existing", user.PasswordHash);
        Assert.Equal(PersistenceConstants.ArticleQueued, article.Status);
        Assert.Equal("https://example.com/article", article.OriginalUrl);
        Assert.Equal(article.Id, result.ArticleId);
        Assert.Equal(PersistenceConstants.JobQueued, job.Status);
        Assert.Equal(PersistenceConstants.ArticleProcessingJobType, job.Type);
        Assert.Equal(100, job.TelegramUpdateId);
        Assert.Equal(200, job.TelegramChatId);
        Assert.Equal(300, job.TelegramMessageId);
        Assert.Equal(400, job.TelegramUserId);
        Assert.Equal(job.Id, result.JobId);
    }

    [Fact]
    public async Task RecordValidUrlIgnoresDuplicateTelegramUpdate()
    {
        await using var db = await CreateDbAsync();
        var repo = CreateRepository(db);
        var command = new RecordTelegramIngestionCommand(100, 200, 300, 400, "https://example.com/article");

        var first = await repo.RecordValidUrlAsync(command, CancellationToken.None);
        var second = await repo.RecordValidUrlAsync(command, CancellationToken.None);

        Assert.True(first.Created);
        Assert.False(second.Created);
        Assert.Equal(first.ArticleId, second.ArticleId);
        Assert.Equal(first.JobId, second.JobId);
        Assert.Equal(1, await db.Articles.CountAsync(CancellationToken.None));
        Assert.Equal(1, await db.Jobs.CountAsync(CancellationToken.None));
    }

    [Fact]
    public async Task SchemaConstrainsCanonicalStatesAndOmitsArtifactPathColumns()
    {
        await using var db = await CreateDbAsync();
        var columns = new List<string>();
        var connection = db.Database.GetDbConnection();
        await connection.OpenAsync(CancellationToken.None);
        await using var command = connection.CreateCommand();
        command.CommandText = "PRAGMA table_info(articles);";
        await using var reader = await command.ExecuteReaderAsync(CancellationToken.None);

        while (await reader.ReadAsync(CancellationToken.None))
        {
            columns.Add(reader.GetString(1));
        }

        Assert.Contains("original_url", columns);
        Assert.DoesNotContain("summary", columns);
        Assert.DoesNotContain("domain", columns);
        Assert.DoesNotContain("artifact_path", columns);
        Assert.DoesNotContain("selected_extractor", columns);
        Assert.DoesNotContain("extractor_score", columns);
        Assert.DoesNotContain("processed_at", columns);
    }

    [Fact]
    public void ArticleArtifactPathsAreDerivedFromDataDirAndArticleId()
    {
        var paths = new ArticleArtifactPaths("/data");
        const string articleId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";

        Assert.Equal(Path.Combine("/data", "articles", articleId), paths.ArticleDirectory(articleId));
        Assert.Equal(Path.Combine("/data", "articles", articleId, "snapshot.html"), paths.SnapshotHtml(articleId));
        Assert.Equal(Path.Combine("/data", "articles", articleId, "content.md"), paths.ContentMarkdown(articleId));
        Assert.Equal(Path.Combine("/data", "articles", articleId, "summary.md"), paths.SummaryMarkdown(articleId));
    }

    private static async Task<ArchivistDbContext> CreateDbAsync()
    {
        var dbPath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");
        var options = new DbContextOptionsBuilder<ArchivistDbContext>()
            .UseSqlite($"Data Source={dbPath}")
            .Options;
        var db = new ArchivistDbContext(options);
        await db.Database.EnsureCreatedAsync(CancellationToken.None);

        return db;
    }

    private static EfTelegramIngestionRepository CreateRepository(ArchivistDbContext db)
    {
        var time = new FakeTimeProvider(new DateTimeOffset(2026, 5, 7, 12, 0, 0, TimeSpan.Zero));

        return new EfTelegramIngestionRepository(db, new SequentialIdGenerator(), time);
    }

    private sealed class SequentialIdGenerator : IUlidGenerator
    {
        private int _next = 1;

        public string NewId() => _next++ switch
        {
            1 => "01ASB2XFCZJY7WHZ2FNRTMQJC1",
            2 => "01ASB2XFCZJY7WHZ2FNRTMQJC2",
            3 => "01ASB2XFCZJY7WHZ2FNRTMQJC3",
            _ => "01ASB2XFCZJY7WHZ2FNRTMQJC4",
        };
    }
}