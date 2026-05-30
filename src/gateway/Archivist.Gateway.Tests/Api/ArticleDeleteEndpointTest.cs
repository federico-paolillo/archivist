namespace Archivist.Gateway.Tests.Api;

using System.Data.Common;
using System.Net;

using Archivist.Gateway.Application.ArticleArtifacts;
using Archivist.Gateway.Application.Articles;
using Archivist.Gateway.Application.Articles.Defaults;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Diagnostics;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Time.Testing;

using Xunit.Abstractions;

public sealed class ArticleDeleteEndpointTest(ITestOutputHelper testOutputHelper) : IntegrationTest(testOutputHelper)
{
    private const string CookieName = "__Host-app-auth";
    private const string SessionId = "test-session-id";
    private const string PublicHost = "localhost";
    private const string PublicOrigin = "https://localhost";
    private const string PersonalUserId = PersistenceConstants.PersonalUserId;
    private const string OtherUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCV";

    [Theory]
    [InlineData(PersistenceConstants.ArticleReady, PersistenceConstants.JobSucceeded)]
    [InlineData(PersistenceConstants.ArticleFailed, PersistenceConstants.JobFailed)]
    [InlineData(PersistenceConstants.ArticleQueued, PersistenceConstants.JobQueued)]
    public async Task DeleteArticle_AllowedStatuses_RemovesArticleRowsAndArtifacts(string articleStatus, string jobStatus)
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000000";
        var jobId = "01H00000000000000000000001";

        await SeedArticleAsync(env.SqlitePath, articleId, PersonalUserId, articleStatus, jobId, jobStatus, addNotification: true);
        var artifactFile = CreateArtifact(env.DataDirectory, articleId);

        using var http = CreateHttpClient();
        var response = await SendDeleteAsync(http, articleId);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
        Assert.False(Directory.Exists(Path.GetDirectoryName(artifactFile)));

        await using var db = CreateDb(env.SqlitePath);
        Assert.Equal(0, await db.Articles.CountAsync());
        Assert.Equal(0, await db.Jobs.CountAsync());
        Assert.Equal(0, await db.Notifications.CountAsync());
    }

    [Fact]
    public async Task DeleteArticle_RunningJob_Returns409AndLeavesState()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000002";

        await SeedArticleAsync(env.SqlitePath, articleId, PersonalUserId, PersistenceConstants.ArticleQueued, "01H00000000000000000000003", PersistenceConstants.JobRunning, addNotification: false);
        CreateArtifact(env.DataDirectory, articleId);

        using var http = CreateHttpClient();
        var response = await SendDeleteAsync(http, articleId);

        Assert.Equal(HttpStatusCode.Conflict, response.StatusCode);
        Assert.True(Directory.Exists(Path.Combine(env.DataDirectory, "articles", articleId)));

        await using var db = CreateDb(env.SqlitePath);
        Assert.Equal(1, await db.Articles.CountAsync());
        Assert.Equal(1, await db.Jobs.CountAsync());
    }

    [Fact]
    public async Task DeleteArticle_MalformedId_Returns400()
    {
        PrepareArticleEnvironment();

        using var http = CreateHttpClient();
        var response = await SendDeleteAsync(http, "not-a-ulid");

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    [Fact]
    public async Task DeleteArticle_NotOwnedByAuthenticatedUser_Returns404()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000004";

        await SeedArticleAsync(env.SqlitePath, articleId, OtherUserId, PersistenceConstants.ArticleQueued, "01H00000000000000000000005", PersistenceConstants.JobQueued, addNotification: false);

        using var http = CreateHttpClient();
        var response = await SendDeleteAsync(http, articleId);

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);

        await using var db = CreateDb(env.SqlitePath);
        Assert.Equal(1, await db.Articles.CountAsync());
        Assert.Equal(1, await db.Jobs.CountAsync());
    }

    [Fact]
    public async Task DeleteArticle_MissingArtifactDirectory_DoesNotFail()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000006";

        await SeedArticleAsync(env.SqlitePath, articleId, PersonalUserId, PersistenceConstants.ArticleQueued, "01H00000000000000000000007", PersistenceConstants.JobQueued, addNotification: false);

        using var http = CreateHttpClient();
        var response = await SendDeleteAsync(http, articleId);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
    }

    [Fact]
    public async Task DeleteArticle_ArtifactCleanupFailure_Returns500AndLeavesDatabaseState()
    {
        var env = PrepareArticleEnvironment(artifactDeletion: new FailingArticleArtifactDeletion());
        var articleId = "01H00000000000000000000008";

        await SeedArticleAsync(env.SqlitePath, articleId, PersonalUserId, PersistenceConstants.ArticleQueued, "01H00000000000000000000009", PersistenceConstants.JobQueued, addNotification: true);

        using var http = CreateHttpClient();
        var response = await SendDeleteAsync(http, articleId);

        Assert.Equal(HttpStatusCode.InternalServerError, response.StatusCode);

        await using var db = CreateDb(env.SqlitePath);
        Assert.Equal(1, await db.Articles.CountAsync());
        Assert.Equal(1, await db.Jobs.CountAsync());
        Assert.Equal(1, await db.Notifications.CountAsync());
    }

    [Fact]
    public async Task DeleteArticle_ArtifactCleanupFailureAfterDatabaseDeletes_RollsBackDatabaseState()
    {
        var sqlitePath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");
        var articleId = "01H00000000000000000000010";
        var jobId = "01H00000000000000000000011";
        var operationLog = new List<string>();

        await SeedArticleAsync(
            sqlitePath,
            articleId,
            PersonalUserId,
            PersistenceConstants.ArticleQueued,
            jobId,
            PersistenceConstants.JobQueued,
            addNotification: true);

        await using (var db = CreateDb(sqlitePath, new DeleteCommandRecorder(operationLog)))
        {
            var service = new EfArticleDeleteService(
                db,
                new RecordingFailingArticleArtifactDeletion(operationLog));

            var result = await service.DeleteAsync(articleId, PersonalUserId, CancellationToken.None);

            Assert.Equal(ArticleDeleteResult.ArtifactCleanupFailed, result);
        }

        Assert.Contains("delete notifications", operationLog);
        Assert.Contains("delete jobs", operationLog);
        Assert.Contains("delete articles", operationLog);
        Assert.Contains("artifact cleanup", operationLog);
        Assert.True(operationLog.IndexOf("delete notifications") < operationLog.IndexOf("artifact cleanup"));
        Assert.True(operationLog.IndexOf("delete jobs") < operationLog.IndexOf("artifact cleanup"));
        Assert.True(operationLog.IndexOf("delete articles") < operationLog.IndexOf("artifact cleanup"));

        await using var verificationDb = CreateDb(sqlitePath);
        Assert.Equal(1, await verificationDb.Articles.CountAsync());
        Assert.Equal(1, await verificationDb.Jobs.CountAsync());
        Assert.Equal(1, await verificationDb.Notifications.CountAsync());
    }

    [Fact]
    public async Task DeleteArticle_CrossSiteOrigin_Returns403()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H0000000000000000000000A";

        await SeedArticleAsync(env.SqlitePath, articleId, PersonalUserId, PersistenceConstants.ArticleQueued, "01H0000000000000000000000B", PersistenceConstants.JobQueued, addNotification: false);

        using var http = CreateHttpClient();
        using var request = CreateDeleteRequest(articleId, "https://evil.example.com");
        AddTrustedForwardedHeaders(request);

        var response = await http.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Theory]
    [InlineData("http://localhost", null)]
    [InlineData("https://example.net", null)]
    [InlineData("https://localhost:9443", "localhost:8443")]
    public async Task DeleteArticle_OriginMismatch_Returns403(string origin, string? host)
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H0000000000000000000000E";

        await SeedArticleAsync(env.SqlitePath, articleId, PersonalUserId, PersistenceConstants.ArticleQueued, "01H0000000000000000000000F", PersistenceConstants.JobQueued, addNotification: false);

        using var http = CreateHttpClient();
        using var request = CreateDeleteRequest(articleId, origin);
        AddTrustedForwardedHeaders(request);

        if (host is not null)
        {
            request.Headers.Host = host;
            request.Headers.Remove("X-Forwarded-Host");
        }

        var response = await http.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Fact]
    public async Task DeleteArticle_QueuedArticle_RemovesQueuedJobSoLaterClaimFindsNothing()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H0000000000000000000000C";

        await SeedArticleAsync(env.SqlitePath, articleId, PersonalUserId, PersistenceConstants.ArticleQueued, "01H0000000000000000000000D", PersistenceConstants.JobQueued, addNotification: false);

        using var http = CreateHttpClient();
        var response = await SendDeleteAsync(http, articleId);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);

        await using var db = CreateDb(env.SqlitePath);
        var claimableJobs = await db.Jobs.CountAsync(x => x.Status == PersistenceConstants.JobQueued);
        Assert.Equal(0, claimableJobs);
    }

    private TestEnvironment PrepareArticleEnvironment(IArticleArtifactDeletion? artifactDeletion = null)
    {
        var sqlitePath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");
        var dataDirectory = Path.Combine(Path.GetTempPath(), $"archivist-data-{Guid.NewGuid():N}");
        var fakeTime = new FakeTimeProvider(DateTimeOffset.UtcNow);
        var sessionStore = new InMemorySessionStore(fakeTime);
        var now = fakeTime.GetUtcNow();
        sessionStore
            .SetAsync(SessionId, new SessionEntry(PersonalUserId, now, now.AddHours(24)))
            .GetAwaiter()
            .GetResult();

        PrepareEnvironment(
            Environments.Development,
            configureTestServices: services =>
            {
                services.RemoveAll<ISessionStore>();
                services.RemoveAll<TimeProvider>();
                services.RemoveAll<IHostedService>();
                services.AddSingleton<ISessionStore>(sessionStore);
                services.AddSingleton<TimeProvider>(fakeTime);

                if (artifactDeletion is not null)
                {
                    services.RemoveAll<IArticleArtifactDeletion>();
                    services.AddScoped(_ => artifactDeletion);
                }
            },
            configureConfiguration: cfg =>
                cfg.AddInMemoryCollection(new Dictionary<string, string?>
                {
                    ["SQLITE_PATH"] = sqlitePath,
                    ["DATA_DIR"] = dataDirectory,
                    ["Telegram:WebhookSecret"] = "test-webhook-secret",
                    ["Telegram:AllowedUserId"] = "99999",
                    ["Telegram:BotToken"] = "fake-token",
                    ["GATEWAY_PUBLIC_HOSTS"] = PublicHost,
                }));

        return new TestEnvironment(sqlitePath, dataDirectory);
    }

    private async Task<HttpResponseMessage> SendDeleteAsync(HttpClient http, string articleId)
    {
        using var request = CreateDeleteRequest(articleId, PublicOrigin);
        AddTrustedForwardedHeaders(request);
        return await http.SendAsync(request);
    }

    private static HttpRequestMessage CreateDeleteRequest(string articleId, string origin)
    {
        var request = new HttpRequestMessage(HttpMethod.Delete, $"/articles/{articleId}");
        request.Headers.Add("Cookie", $"{CookieName}={SessionId}");
        request.Headers.Add("Origin", origin);
        return request;
    }

    private static void AddTrustedForwardedHeaders(HttpRequestMessage request)
    {
        request.Headers.Add("X-Forwarded-Proto", "https");
        request.Headers.Add("X-Forwarded-For", "203.0.113.20");
        request.Headers.Add("X-Forwarded-Host", PublicHost);
    }

    private static string CreateArtifact(string dataDirectory, string articleId)
    {
        var articleDirectory = Path.Combine(dataDirectory, "articles", articleId);
        Directory.CreateDirectory(articleDirectory);
        var artifactFile = Path.Combine(articleDirectory, "summary.md");
        File.WriteAllText(artifactFile, "summary");
        return artifactFile;
    }

    private static async Task SeedArticleAsync(
        string sqlitePath,
        string articleId,
        string userId,
        string articleStatus,
        string jobId,
        string jobStatus,
        bool addNotification)
    {
        await using var db = CreateDb(sqlitePath);
        await db.Database.EnsureCreatedAsync();

        db.Users.Add(new UserEntity
        {
            Id = userId,
            TelegramUserId = userId == PersonalUserId ? 99999 : null,
        });
        db.Articles.Add(new ArticleEntity
        {
            Id = articleId,
            UserId = userId,
            OriginalUrl = "https://example.com/article",
            Status = articleStatus,
            CreatedAt = DateTimeOffset.UtcNow,
        });
        db.Jobs.Add(new JobEntity
        {
            Id = jobId,
            UserId = userId,
            ArticleId = articleId,
            Type = PersistenceConstants.ArticleProcessingJobType,
            Status = jobStatus,
            TelegramUpdateId = 123456789,
            TelegramChatId = 123,
            TelegramMessageId = 456,
            TelegramUserId = 99999,
            CreatedAt = DateTimeOffset.UtcNow,
        });

        if (addNotification)
        {
            db.Notifications.Add(new NotificationEntity
            {
                Id = $"{jobId[..24]}NF",
                JobId = jobId,
                Status = PersistenceConstants.NotificationPending,
                CreatedAt = DateTimeOffset.UtcNow,
                ExpiresAt = DateTimeOffset.UtcNow.AddDays(7),
            });
        }

        await db.SaveChangesAsync();
    }

    private static ArchivistDbContext CreateDb(string sqlitePath, IInterceptor? interceptor = null)
    {
        var options = new DbContextOptionsBuilder<ArchivistDbContext>()
            .UseSqlite($"Data Source={sqlitePath}");

        if (interceptor is not null)
        {
            options.AddInterceptors(interceptor);
        }

        return new ArchivistDbContext(options.Options);
    }

    private sealed record TestEnvironment(string SqlitePath, string DataDirectory);

    private sealed class FailingArticleArtifactDeletion : IArticleArtifactDeletion
    {
        public Task<bool> DeleteArticleDirectoryAsync(string articleId, CancellationToken cancellationToken) =>
            Task.FromResult(false);
    }

    private sealed class RecordingFailingArticleArtifactDeletion(List<string> operationLog) : IArticleArtifactDeletion
    {
        public Task<bool> DeleteArticleDirectoryAsync(string articleId, CancellationToken cancellationToken)
        {
            operationLog.Add("artifact cleanup");
            return Task.FromResult(false);
        }
    }

    private sealed class DeleteCommandRecorder(List<string> operationLog) : DbCommandInterceptor
    {
        public override ValueTask<InterceptionResult<int>> NonQueryExecutingAsync(
            DbCommand command,
            CommandEventData eventData,
            InterceptionResult<int> result,
            CancellationToken cancellationToken = default)
        {
            Record(command.CommandText);
            return base.NonQueryExecutingAsync(command, eventData, result, cancellationToken);
        }

        private void Record(string commandText)
        {
            if (commandText.Contains("DELETE FROM \"notifications\"", StringComparison.Ordinal))
            {
                operationLog.Add("delete notifications");
            }
            else if (commandText.Contains("DELETE FROM \"jobs\"", StringComparison.Ordinal))
            {
                operationLog.Add("delete jobs");
            }
            else if (commandText.Contains("DELETE FROM \"articles\"", StringComparison.Ordinal))
            {
                operationLog.Add("delete articles");
            }
        }
    }
}