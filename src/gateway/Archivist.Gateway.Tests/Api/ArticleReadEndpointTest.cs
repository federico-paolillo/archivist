namespace Archivist.Gateway.Tests.Api;

using System.Net;
using System.Text.Json;

using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Entities;

using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Time.Testing;

using Xunit.Abstractions;

public sealed class ArticleReadEndpointTest(ITestOutputHelper testOutputHelper) : IntegrationTest(testOutputHelper)
{
    private const string CookieName = "__Host-app-auth";
    private const string SessionId = "test-session-id";
    private const string PublicHost = "localhost";
    private const string PersonalUserId = PersistenceConstants.PersonalUserId;
    private const string OtherUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCV";
    private const string UlidAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ";
    private static readonly DateTimeOffset FixedNow = new(2026, 5, 7, 12, 0, 0, TimeSpan.Zero);

    [Fact]
    public async Task ListArticles_Unauthenticated_Returns401()
    {
        PrepareArticleEnvironment(authenticated: false);

        using var http = CreateHttpClient();
        var response = await http.GetAsync("/articles");

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task ListArticles_FirstPage_ReturnsNewest25AndCursors()
    {
        var env = PrepareArticleEnvironment();
        var articleIds = CreateArticleIds(30);
        await SeedArticlesAsync(env.SqlitePath, articleIds, PersonalUserId, PersistenceConstants.ArticleReady);

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync("/articles");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        using var body = await ParseJsonAsync(response);
        var ids = ReadItemIds(body);
        var expected = articleIds.OrderDescending(StringComparer.Ordinal).Take(25).ToList();

        Assert.Equal(expected, ids);
        Assert.Equal(expected[^1], body.RootElement.GetProperty("nextCursor").GetString());
        Assert.Equal(JsonValueKind.Null, body.RootElement.GetProperty("previousCursor").ValueKind);
    }

    [Fact]
    public async Task ListArticles_AfterCursor_ReturnsOlderArticles()
    {
        var env = PrepareArticleEnvironment();
        var articleIds = CreateArticleIds(30);
        await SeedArticlesAsync(env.SqlitePath, articleIds, PersonalUserId, PersistenceConstants.ArticleReady);
        var ordered = articleIds.OrderDescending(StringComparer.Ordinal).ToList();

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles?after={ordered[24]}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        using var body = await ParseJsonAsync(response);
        var ids = ReadItemIds(body);
        var expected = ordered.Skip(25).ToList();

        Assert.Equal(expected, ids);
        Assert.Equal(JsonValueKind.Null, body.RootElement.GetProperty("nextCursor").ValueKind);
        Assert.Equal(expected[0], body.RootElement.GetProperty("previousCursor").GetString());
    }

    [Fact]
    public async Task ListArticles_BeforeCursor_ReturnsNewerArticles()
    {
        var env = PrepareArticleEnvironment();
        var articleIds = CreateArticleIds(30);
        await SeedArticlesAsync(env.SqlitePath, articleIds, PersonalUserId, PersistenceConstants.ArticleReady);
        var ordered = articleIds.OrderDescending(StringComparer.Ordinal).ToList();

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles?before={ordered[25]}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        using var body = await ParseJsonAsync(response);
        var ids = ReadItemIds(body);
        var expected = ordered.Take(25).ToList();

        Assert.Equal(expected, ids);
        Assert.Equal(expected[^1], body.RootElement.GetProperty("nextCursor").GetString());
        Assert.Equal(JsonValueKind.Null, body.RootElement.GetProperty("previousCursor").ValueKind);
    }

    [Theory]
    [InlineData("/articles?after=not-a-ulid")]
    [InlineData("/articles?before=not-a-ulid")]
    [InlineData("/articles?after=01H00000000000000000000000&before=01H00000000000000000000001")]
    public async Task ListArticles_InvalidCursors_Returns400(string requestPath)
    {
        PrepareArticleEnvironment();

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync(requestPath);

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    [Fact]
    public async Task GetArticle_Unauthenticated_Returns401()
    {
        PrepareArticleEnvironment(authenticated: false);

        using var http = CreateHttpClient();
        var response = await http.GetAsync("/articles/01H00000000000000000000000");

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task GetArticle_MalformedId_Returns400()
    {
        PrepareArticleEnvironment();

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync("/articles/not-a-ulid");

        Assert.Equal(HttpStatusCode.BadRequest, response.StatusCode);
    }

    [Fact]
    public async Task GetArticle_NotOwnedByAuthenticatedUser_Returns404()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000000";
        await SeedArticlesAsync(env.SqlitePath, new List<string> { articleId }, OtherUserId, PersistenceConstants.ArticleReady);

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles/{articleId}");

        Assert.Equal(HttpStatusCode.NotFound, response.StatusCode);
    }

    [Fact]
    public async Task GetArticle_ReadyArticle_ReturnsMarkdownArtifacts()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000001";
        await SeedArticlesAsync(env.SqlitePath, new List<string> { articleId }, PersonalUserId, PersistenceConstants.ArticleReady);
        WriteArtifacts(env.DataDirectory, articleId, summaryMarkdown: "Summary text", contentMarkdown: "Content text");

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles/{articleId}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        using var body = await ParseJsonAsync(response);
        var root = body.RootElement;
        Assert.Equal(articleId, root.GetProperty("id").GetString());
        Assert.Equal("Summary text", root.GetProperty("summaryMarkdown").GetString());
        Assert.Equal("Content text", root.GetProperty("contentMarkdown").GetString());
        Assert.False(root.GetProperty("canForceDelete").GetBoolean());
    }

    [Fact]
    public async Task GetArticle_ReadyArticleMissingRequiredArtifact_Returns500()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000002";
        await SeedArticlesAsync(env.SqlitePath, new List<string> { articleId }, PersonalUserId, PersistenceConstants.ArticleReady);
        WriteArtifacts(env.DataDirectory, articleId, summaryMarkdown: "Summary text", contentMarkdown: null);

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles/{articleId}");

        Assert.Equal(HttpStatusCode.InternalServerError, response.StatusCode);
    }

    [Theory]
    [InlineData(PersistenceConstants.ArticleQueued)]
    [InlineData(PersistenceConstants.ArticleFailed)]
    public async Task GetArticle_NonReadyArticleMissingArtifacts_ReturnsNullMarkdown(string status)
    {
        var env = PrepareArticleEnvironment();
        var articleId = status == PersistenceConstants.ArticleQueued
            ? "01H00000000000000000000003"
            : "01H00000000000000000000004";
        await SeedArticlesAsync(env.SqlitePath, new List<string> { articleId }, PersonalUserId, status);

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles/{articleId}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        using var body = await ParseJsonAsync(response);
        var root = body.RootElement;
        Assert.Equal(JsonValueKind.Null, root.GetProperty("summaryMarkdown").ValueKind);
        Assert.Equal(JsonValueKind.Null, root.GetProperty("contentMarkdown").ValueKind);
        Assert.False(root.GetProperty("canForceDelete").GetBoolean());
    }

    [Fact]
    public async Task GetArticle_StaleRunningJob_ReturnsCanForceDeleteTrue()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000005";
        await SeedArticlesAsync(env.SqlitePath, new List<string> { articleId }, PersonalUserId, PersistenceConstants.ArticleQueued);
        await SeedJobAsync(
            env.SqlitePath,
            articleId,
            "01H00000000000000000000006",
            PersonalUserId,
            PersistenceConstants.JobRunning,
            env.Now.AddHours(-3));

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles/{articleId}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        using var body = await ParseJsonAsync(response);
        Assert.True(body.RootElement.GetProperty("canForceDelete").GetBoolean());
    }

    [Fact]
    public async Task GetArticle_ActiveRunningJob_ReturnsCanForceDeleteFalse()
    {
        var env = PrepareArticleEnvironment();
        var articleId = "01H00000000000000000000007";
        await SeedArticlesAsync(env.SqlitePath, new List<string> { articleId }, PersonalUserId, PersistenceConstants.ArticleQueued);
        await SeedJobAsync(
            env.SqlitePath,
            articleId,
            "01H00000000000000000000008",
            PersonalUserId,
            PersistenceConstants.JobRunning,
            env.Now.AddHours(-1));

        using var http = CreateAuthenticatedHttpClient();
        var response = await http.GetAsync($"/articles/{articleId}");

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);

        using var body = await ParseJsonAsync(response);
        Assert.False(body.RootElement.GetProperty("canForceDelete").GetBoolean());
    }

    private TestEnvironment PrepareArticleEnvironment(bool authenticated = true)
    {
        var sqlitePath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");
        var dataDirectory = Path.Combine(Path.GetTempPath(), $"archivist-data-{Guid.NewGuid():N}");
        var fakeTime = new FakeTimeProvider(FixedNow);
        var sessionStore = new InMemorySessionStore(fakeTime);

        if (authenticated)
        {
            var now = fakeTime.GetUtcNow();
            sessionStore
                .SetAsync(SessionId, new SessionEntry(PersonalUserId, now, now.AddHours(24)))
                .GetAwaiter()
                .GetResult();
        }

        PrepareEnvironment(
            Environments.Development,
            configureTestServices: services =>
            {
                services.RemoveAll<ISessionStore>();
                services.RemoveAll<TimeProvider>();
                services.RemoveAll<IHostedService>();
                services.AddSingleton<ISessionStore>(sessionStore);
                services.AddSingleton<TimeProvider>(fakeTime);
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

        return new TestEnvironment(sqlitePath, dataDirectory, fakeTime.GetUtcNow());
    }

    private HttpClient CreateAuthenticatedHttpClient()
    {
        var http = CreateHttpClient();
        http.DefaultRequestHeaders.Add("Cookie", $"{CookieName}={SessionId}");
        return http;
    }

    private static async Task<JsonDocument> ParseJsonAsync(HttpResponseMessage response)
    {
        await using var stream = await response.Content.ReadAsStreamAsync();
        return await JsonDocument.ParseAsync(stream);
    }

    private static List<string> ReadItemIds(JsonDocument body) =>
        body.RootElement
            .GetProperty("items")
            .EnumerateArray()
            .Select(item => item.GetProperty("id").GetString() ?? string.Empty)
            .ToList();

    private static List<string> CreateArticleIds(int count) =>
        Enumerable.Range(0, count)
            .Select(CreateArticleId)
            .ToList();

    private static string CreateArticleId(int value)
    {
        var high = value / UlidAlphabet.Length;
        var low = value % UlidAlphabet.Length;
        return $"01H000000000000000000000{UlidAlphabet[high]}{UlidAlphabet[low]}";
    }

    private static void WriteArtifacts(
        string dataDirectory,
        string articleId,
        string? summaryMarkdown,
        string? contentMarkdown)
    {
        var articleDirectory = Path.Combine(dataDirectory, "articles", articleId);
        Directory.CreateDirectory(articleDirectory);

        if (summaryMarkdown is not null)
        {
            File.WriteAllText(Path.Combine(articleDirectory, "summary.md"), summaryMarkdown);
        }

        if (contentMarkdown is not null)
        {
            File.WriteAllText(Path.Combine(articleDirectory, "content.md"), contentMarkdown);
        }
    }

    private static async Task SeedArticlesAsync(
        string sqlitePath,
        List<string> articleIds,
        string userId,
        string articleStatus)
    {
        await using var db = CreateDb(sqlitePath);
        await db.Database.EnsureCreatedAsync();

        if (!await db.Users.AnyAsync(x => x.Id == userId))
        {
            db.Users.Add(new UserEntity
            {
                Id = userId,
                TelegramUserId = userId == PersonalUserId ? 99999 : null,
            });
        }

        for (var index = 0; index < articleIds.Count; index++)
        {
            var articleId = articleIds[index];

            db.Articles.Add(new ArticleEntity
            {
                Id = articleId,
                UserId = userId,
                OriginalUrl = $"https://example.com/articles/{articleId}",
                CanonicalUrl = $"https://canonical.example.com/articles/{articleId}",
                Title = $"Article {articleId}",
                Status = articleStatus,
                ErrorMessage = articleStatus == PersistenceConstants.ArticleFailed ? "[ARC-013] Summary failed." : null,
                CreatedAt = DateTimeOffset.UtcNow.AddMinutes(index),
            });
        }

        await db.SaveChangesAsync();
    }

    private static async Task SeedJobAsync(
        string sqlitePath,
        string articleId,
        string jobId,
        string userId,
        string status,
        DateTimeOffset? startedAt)
    {
        await using var db = CreateDb(sqlitePath);
        db.Jobs.Add(new JobEntity
        {
            Id = jobId,
            UserId = userId,
            ArticleId = articleId,
            Type = PersistenceConstants.ArticleProcessingJobType,
            Status = status,
            CreatedAt = FixedNow,
            StartedAt = startedAt,
        });

        await db.SaveChangesAsync();
    }

    private static ArchivistDbContext CreateDb(string sqlitePath)
    {
        var options = new DbContextOptionsBuilder<ArchivistDbContext>()
            .UseSqlite($"Data Source={sqlitePath}")
            .Options;

        return new ArchivistDbContext(options);
    }

    private sealed record TestEnvironment(string SqlitePath, string DataDirectory, DateTimeOffset Now);
}