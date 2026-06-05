using System.Data.Common;
using System.Diagnostics;

using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.Observability;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Defaults;
using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Diagnostics;
using Microsoft.Extensions.Time.Testing;

namespace Archivist.Gateway.Tests.Persistence;

public sealed class TelegramIngestionRepositoryTest
{
    private const string TestUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";
    private const string OtherUserId = "01BSB2XFCZJY7WHZ2FNRTMQJCT";

    [Fact]
    public async Task RecordValidUrlCreatesArticleAndQueuedJobForResolvedUser()
    {
        await using var db = await CreateDbAsync();
        db.Users.Add(new UserEntity
        {
            Id = TestUserId,
            TelegramUserId = 400,
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
                TestUserId,
                "https://example.com/article"),
            CancellationToken.None);

        Assert.True(result.Created);

        var user = await db.Users.SingleAsync(CancellationToken.None);
        var article = await db.Articles.SingleAsync(CancellationToken.None);
        var job = await db.Jobs.SingleAsync(CancellationToken.None);

        Assert.Equal(TestUserId, user.Id);
        Assert.Equal(400, user.TelegramUserId);
        Assert.Equal("$argon2id$existing", user.PasswordHash);
        Assert.Equal(TestUserId, article.UserId);
        Assert.Equal(PersistenceConstants.ArticleQueued, article.Status);
        Assert.Equal("https://example.com/article", article.OriginalUrl);
        Assert.Equal(article.Id, result.ArticleId);
        Assert.Equal(TestUserId, job.UserId);
        Assert.Equal(PersistenceConstants.JobQueued, job.Status);
        Assert.Equal(PersistenceConstants.ArticleProcessingJobType, job.Type);
        Assert.Equal(100, job.TelegramUpdateId);
        Assert.Equal(200, job.TelegramChatId);
        Assert.Equal(300, job.TelegramMessageId);
        Assert.Equal(400, job.TelegramUserId);
        Assert.Equal(job.Id, result.JobId);
    }

    [Fact]
    public async Task RecordValidUrlPersistsTraceCarrier()
    {
        await using var db = await CreateDbAsync();
        db.Users.Add(new UserEntity { Id = TestUserId, TelegramUserId = 400 });
        await db.SaveChangesAsync();
        var repo = CreateRepository(db);

        await repo.RecordValidUrlAsync(
            new RecordTelegramIngestionCommand(
                100,
                200,
                300,
                400,
                TestUserId,
                "https://example.com/article",
                "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
                "vendor=value"),
            CancellationToken.None);

        var job = await db.Jobs.SingleAsync(CancellationToken.None);

        Assert.Equal("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", job.TraceParent);
        Assert.Equal("vendor=value", job.TraceState);
    }

    [Fact]
    public async Task RecordValidUrlActivityIncludesUserId()
    {
        await using var db = await CreateDbAsync();
        db.Users.Add(new UserEntity { Id = TestUserId, TelegramUserId = 400 });
        await db.SaveChangesAsync();
        var repo = CreateRepository(db);
        var stoppedActivities = new List<Activity>();
        using var listener = new ActivityListener
        {
            ShouldListenTo = static source => source.Name == ArchivistTelemetry.ActivitySourceName,
            Sample = static (ref ActivityCreationOptions<ActivityContext> _) => ActivitySamplingResult.AllDataAndRecorded,
            ActivityStopped = stoppedActivities.Add,
        };
        ActivitySource.AddActivityListener(listener);

        await repo.RecordValidUrlAsync(
            new RecordTelegramIngestionCommand(100, 200, 300, 400, TestUserId, "https://example.com/article"),
            CancellationToken.None);

        var activity = Assert.Single(stoppedActivities, x => x.DisplayName == "gateway.telegram.enqueue");
        var userIdTag = Assert.Single(activity.Tags, x => x.Key == "user_id");
        Assert.Equal(TestUserId, userIdTag.Value);
    }

    [Fact]
    public async Task GatewaySchemaUpgraderAddsTraceCarrierColumnsIdempotently()
    {
        var dbPath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");
        await using (var connection = new SqliteConnection($"Data Source={dbPath}"))
        {
            await connection.OpenAsync(CancellationToken.None);
            await using var command = connection.CreateCommand();
            command.CommandText = "CREATE TABLE jobs (id TEXT PRIMARY KEY);";
            await command.ExecuteNonQueryAsync(CancellationToken.None);
        }

        await using var db = CreateDbContext(dbPath);

        await GatewaySchemaUpgrader.EnsureJobTraceCarrierColumnsAsync(db, CancellationToken.None);
        await GatewaySchemaUpgrader.EnsureJobTraceCarrierColumnsAsync(db, CancellationToken.None);

        var columns = await ReadJobsTableColumnsAsync(db);

        Assert.Contains("traceparent", columns);
        Assert.Contains("tracestate", columns);
    }

    [Fact]
    public async Task RecordValidUrlIgnoresDuplicateTelegramUpdate()
    {
        await using var db = await CreateDbAsync();
        db.Users.Add(new UserEntity { Id = TestUserId, TelegramUserId = 400 });
        await db.SaveChangesAsync();
        var repo = CreateRepository(db);
        var command = new RecordTelegramIngestionCommand(100, 200, 300, 400, TestUserId, "https://example.com/article");

        var first = await repo.RecordValidUrlAsync(command, CancellationToken.None);
        var second = await repo.RecordValidUrlAsync(command, CancellationToken.None);

        Assert.True(first.Created);
        Assert.False(second.Created);
        Assert.Equal(first.ArticleId, second.ArticleId);
        Assert.Equal(first.JobId, second.JobId);
        Assert.Equal(1, await db.Articles.CountAsync(CancellationToken.None));
        Assert.Equal(1, await db.Jobs.CountAsync(CancellationToken.None));

        var user = await db.Users.SingleAsync(CancellationToken.None);
        Assert.Equal(command.TelegramUserId, user.TelegramUserId);
        Assert.Equal(TestUserId, user.Id);
    }

    [Fact]
    public async Task RecordValidUrlDoesNotCreateOrReassignUserRows()
    {
        await using var db = await CreateDbAsync();
        db.Users.AddRange(
            new UserEntity
            {
                Id = TestUserId,
                TelegramUserId = 400,
                PasswordHash = "$argon2id$existing",
            },
            new UserEntity
            {
                Id = OtherUserId,
                TelegramUserId = 401,
                PasswordHash = null,
            });
        await db.SaveChangesAsync();
        var repo = CreateRepository(db);

        await repo.RecordValidUrlAsync(
            new RecordTelegramIngestionCommand(100, 200, 300, 400, TestUserId, "https://example.com/article"),
            CancellationToken.None);

        var users = await db.Users.OrderBy(x => x.Id).ToListAsync(CancellationToken.None);
        var article = await db.Articles.SingleAsync(CancellationToken.None);
        var job = await db.Jobs.SingleAsync(CancellationToken.None);

        Assert.Equal(2, users.Count);
        Assert.Equal(TestUserId, article.UserId);
        Assert.Equal(TestUserId, job.UserId);
        Assert.Equal("$argon2id$existing", users.Single(x => x.Id == TestUserId).PasswordHash);
        Assert.Equal(400, users.Single(x => x.Id == TestUserId).TelegramUserId);
        Assert.Equal(401, users.Single(x => x.Id == OtherUserId).TelegramUserId);
    }

    [Fact]
    public async Task RecordValidUrlConcurrentDuplicateTelegramUpdateReturnsExistingJob()
    {
        var dbPath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");
        await EnsureSchemaAsync(dbPath);
        await using (var seedDb = CreateDbContext(dbPath))
        {
            seedDb.Users.Add(new UserEntity { Id = TestUserId, TelegramUserId = 400 });
            await seedDb.SaveChangesAsync();
        }

        var barrier = new TransactionStartBarrier();
        await using var db1 = CreateDbContext(dbPath, barrier);
        await using var db2 = CreateDbContext(dbPath, barrier);
        var time = new FakeTimeProvider(new DateTimeOffset(2026, 5, 7, 12, 0, 0, TimeSpan.Zero));
        var repo1 = new EfTelegramIngestionRepository(
            db1,
            new QueueIdGenerator("01H00000000000000000000012", "01H00000000000000000000013"),
            time);
        var repo2 = new EfTelegramIngestionRepository(
            db2,
            new QueueIdGenerator("01H00000000000000000000014", "01H00000000000000000000015"),
            time);
        var command = new RecordTelegramIngestionCommand(100, 200, 300, 400, TestUserId, "https://example.com/article");
        var start = new TaskCompletionSource(TaskCreationOptions.RunContinuationsAsynchronously);
        var task1 = Task.Run(async () =>
        {
            await start.Task;
            return await repo1.RecordValidUrlAsync(command, CancellationToken.None);
        });
        var task2 = Task.Run(async () =>
        {
            await start.Task;
            return await repo2.RecordValidUrlAsync(command, CancellationToken.None);
        });

        start.SetResult();
        var results = await Task.WhenAll(task1, task2);

        var created = Assert.Single(results, x => x.Created);
        var duplicate = Assert.Single(results, x => !x.Created);
        Assert.Equal(created.ArticleId, duplicate.ArticleId);
        Assert.Equal(created.JobId, duplicate.JobId);

        await using var verificationDb = CreateDbContext(dbPath);
        Assert.Equal(1, await verificationDb.Articles.CountAsync(CancellationToken.None));
        Assert.Equal(1, await verificationDb.Jobs.CountAsync(CancellationToken.None));

        var job = await verificationDb.Jobs.SingleAsync(CancellationToken.None);
        Assert.Equal(command.TelegramUpdateId, job.TelegramUpdateId);
        Assert.Equal(created.ArticleId, job.ArticleId);
        Assert.Equal(created.JobId, job.Id);
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

    private static async Task<ArchivistDbContext> CreateDbAsync(string? dbPath = null)
    {
        dbPath ??= Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");
        var db = CreateDbContext(dbPath);
        await db.Database.EnsureCreatedAsync(CancellationToken.None);

        return db;
    }

    private static ArchivistDbContext CreateDbContext(string dbPath, IInterceptor? interceptor = null)
    {
        var options = new DbContextOptionsBuilder<ArchivistDbContext>()
            .UseSqlite($"Data Source={dbPath}");

        if (interceptor is not null)
        {
            options.AddInterceptors(interceptor);
        }

        return new ArchivistDbContext(options.Options);
    }

    private static async Task EnsureSchemaAsync(string dbPath)
    {
        await using var db = CreateDbContext(dbPath);
        await db.Database.EnsureCreatedAsync(CancellationToken.None);
    }

    private static async Task<List<string>> ReadJobsTableColumnsAsync(ArchivistDbContext db)
    {
        var columns = new List<string>();
        var connection = db.Database.GetDbConnection();
        await connection.OpenAsync(CancellationToken.None);
        await using var command = connection.CreateCommand();
        command.CommandText = "PRAGMA table_info(jobs);";
        await using var reader = await command.ExecuteReaderAsync(CancellationToken.None);

        while (await reader.ReadAsync(CancellationToken.None))
        {
            columns.Add(reader.GetString(1));
        }

        return columns;
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

    private sealed class QueueIdGenerator(params string[] ids) : IUlidGenerator
    {
        private readonly Queue<string> _ids = new(ids);

        public string NewId()
        {
            return _ids.Count > 0
                ? _ids.Dequeue()
                : throw new InvalidOperationException("The test ID generator is exhausted.");
        }
    }

    private sealed class TransactionStartBarrier : DbTransactionInterceptor
    {
        private const int ExpectedTransactions = 2;
        private readonly TaskCompletionSource _bothTransactionsReached = new(TaskCreationOptions.RunContinuationsAsynchronously);
        private int _remainingTransactions = ExpectedTransactions;
        private int _waitingTransactions;

        public override async ValueTask<InterceptionResult<DbTransaction>> TransactionStartingAsync(
            DbConnection connection,
            TransactionStartingEventData eventData,
            InterceptionResult<DbTransaction> result,
            CancellationToken cancellationToken = default)
        {
            if (Interlocked.Decrement(ref _remainingTransactions) >= 0)
            {
                if (Interlocked.Increment(ref _waitingTransactions) == ExpectedTransactions)
                {
                    _bothTransactionsReached.SetResult();
                }

                await _bothTransactionsReached.Task.WaitAsync(TimeSpan.FromSeconds(5), cancellationToken).ConfigureAwait(false);
            }

            return await base.TransactionStartingAsync(connection, eventData, result, cancellationToken).ConfigureAwait(false);
        }
    }
}