namespace Archivist.Gateway.Tests.Telegram;

using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Defaults;
using Archivist.Gateway.Application.Persistence.Entities;
using Archivist.Gateway.Application.Telegram;

using Microsoft.Data.Sqlite;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Time.Testing;

public sealed class TelegramNotificationDispatcherTest
{
    private const long ChatId = 100;
    private const long MessageId = 200;
    private static readonly DateTimeOffset FixedNow = new(2026, 5, 10, 12, 0, 0, TimeSpan.Zero);

    // -------------------------------------------------------------------------
    // DispatchPendingAsync — succeeded job is deferred
    // -------------------------------------------------------------------------

    [Fact]
    public async Task DispatchPending_SucceededJob_LeavesNotificationPending()
    {
        await using var db = await CreateDbAsync();
        var (jobId, notificationId) = await SeedJobAndNotificationAsync(db.Context, PersistenceConstants.JobSucceeded, errorMessage: null);

        var client = new FakeTelegramClient();
        var dispatcher = CreateDispatcher(db.Context, client);

        await dispatcher.DispatchPendingAsync(CancellationToken.None);

        Assert.Empty(client.SentReplies);

        var notification = await db.Notifications.SingleAsync(CancellationToken.None);
        Assert.Equal(PersistenceConstants.NotificationPending, notification.Status);
    }

    // -------------------------------------------------------------------------
    // DispatchPendingAsync — failed job sends error_message reply
    // -------------------------------------------------------------------------

    [Fact]
    public async Task DispatchPending_FailedJob_SendsErrorMessageAndMarksNotificationSent()
    {
        await using var db = await CreateDbAsync();
        var (jobId, notificationId) = await SeedJobAndNotificationAsync(db.Context, PersistenceConstants.JobFailed, errorMessage: "Something went wrong");

        var client = new FakeTelegramClient();
        var dispatcher = CreateDispatcher(db.Context, client);

        await dispatcher.DispatchPendingAsync(CancellationToken.None);

        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(ChatId, reply.ChatId);
        Assert.Equal(MessageId, reply.ReplyToMessageId);
        Assert.Equal("Something went wrong", reply.Text);

        var notification = await db.Notifications.SingleAsync(CancellationToken.None);
        Assert.Equal(PersistenceConstants.NotificationSent, notification.Status);
        Assert.Equal(FixedNow, notification.SentAt);
        Assert.Equal(FixedNow.AddDays(7), notification.ExpiresAt);
    }

    // -------------------------------------------------------------------------
    // DispatchPendingAsync — ARC-coded error is preserved verbatim
    // -------------------------------------------------------------------------

    [Fact]
    public async Task DispatchPending_ArcCodedError_PreservesErrorTextVerbatim()
    {
        const string arcError = "[ARC-003] The URL was not found.";

        await using var db = await CreateDbAsync();
        await SeedJobAndNotificationAsync(db.Context, PersistenceConstants.JobFailed, errorMessage: arcError);

        var client = new FakeTelegramClient();
        var dispatcher = CreateDispatcher(db.Context, client);

        await dispatcher.DispatchPendingAsync(CancellationToken.None);

        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(arcError, reply.Text);
    }

    // -------------------------------------------------------------------------
    // DispatchPendingAsync — Telegram delivery failure marks notification failed
    // -------------------------------------------------------------------------

    [Fact]
    public async Task DispatchPending_TelegramDeliveryFails_MarksNotificationFailedWithError()
    {
        await using var db = await CreateDbAsync();
        await SeedJobAndNotificationAsync(db.Context, PersistenceConstants.JobFailed, errorMessage: "job error");

        var client = new FakeTelegramClient(failSend: true);
        var dispatcher = CreateDispatcher(db.Context, client);

        await dispatcher.DispatchPendingAsync(CancellationToken.None);

        var notification = await db.Notifications.SingleAsync(CancellationToken.None);
        Assert.Equal(PersistenceConstants.NotificationFailed, notification.Status);
        Assert.NotNull(notification.ErrorMessage);
        Assert.Contains("Telegram delivery failed", notification.ErrorMessage);
        Assert.Equal(FixedNow.AddDays(7), notification.ExpiresAt);
    }

    [Fact]
    public async Task DispatchPending_TelegramDeliveryFails_DoesNotMutateJobOrArticleState()
    {
        await using var db = await CreateDbAsync();
        var (jobId, _) = await SeedJobAndNotificationAsync(db.Context, PersistenceConstants.JobFailed, errorMessage: "job error");

        var client = new FakeTelegramClient(failSend: true);
        var dispatcher = CreateDispatcher(db.Context, client);

        await dispatcher.DispatchPendingAsync(CancellationToken.None);

        var job = await db.Jobs.SingleAsync(j => j.Id == jobId, CancellationToken.None);
        Assert.Equal(PersistenceConstants.JobFailed, job.Status);

        var article = await db.Articles.SingleAsync(CancellationToken.None);
        Assert.Equal(PersistenceConstants.ArticleFailed, article.Status);
    }

    // -------------------------------------------------------------------------
    // Truncation
    // -------------------------------------------------------------------------

    [Fact]
    public void Truncate_ShortMessage_ReturnsUnchanged()
    {
        const string text = "Hello";
        Assert.Equal(text, TelegramNotificationDispatcher.Truncate(text));
    }

    [Fact]
    public void Truncate_ExactlyAtLimit_ReturnsUnchanged()
    {
        var text = new string('a', TelegramNotificationDispatcher.TelegramMessageMaxLength);
        Assert.Equal(text, TelegramNotificationDispatcher.Truncate(text));
    }

    [Fact]
    public void Truncate_ExceedsLimit_TruncatesToLimitWithEllipsis()
    {
        var text = new string('a', TelegramNotificationDispatcher.TelegramMessageMaxLength + 100);
        var result = TelegramNotificationDispatcher.Truncate(text);
        Assert.Equal(TelegramNotificationDispatcher.TelegramMessageMaxLength, result.Length);
        Assert.EndsWith("…", result);
    }

    [Fact]
    public async Task DispatchPending_VeryLongErrorMessage_TruncatesToTelegramLimit()
    {
        var longError = new string('x', TelegramNotificationDispatcher.TelegramMessageMaxLength + 500);

        await using var db = await CreateDbAsync();
        await SeedJobAndNotificationAsync(db.Context, PersistenceConstants.JobFailed, errorMessage: longError);

        var client = new FakeTelegramClient();
        var dispatcher = CreateDispatcher(db.Context, client);

        await dispatcher.DispatchPendingAsync(CancellationToken.None);

        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(TelegramNotificationDispatcher.TelegramMessageMaxLength, reply.Text.Length);
        Assert.EndsWith("…", reply.Text);
    }

    // -------------------------------------------------------------------------
    // CleanUpExpiredAsync — deletes sent and failed notifications past TTL
    // -------------------------------------------------------------------------

    [Fact]
    public async Task CleanUpExpired_SentNotificationPastTtl_IsDeleted()
    {
        await using var db = await CreateDbAsync();
        await SeedTerminalNotificationAsync(db.Context, PersistenceConstants.NotificationSent, expiresAt: FixedNow.AddDays(-1));

        var dispatcher = CreateDispatcher(db.Context, new FakeTelegramClient());
        await dispatcher.CleanUpExpiredAsync(CancellationToken.None);

        Assert.Equal(0, await db.Notifications.CountAsync(CancellationToken.None));
    }

    [Fact]
    public async Task CleanUpExpired_FailedNotificationPastTtl_IsDeleted()
    {
        await using var db = await CreateDbAsync();
        await SeedTerminalNotificationAsync(db.Context, PersistenceConstants.NotificationFailed, expiresAt: FixedNow.AddDays(-1));

        var dispatcher = CreateDispatcher(db.Context, new FakeTelegramClient());
        await dispatcher.CleanUpExpiredAsync(CancellationToken.None);

        Assert.Equal(0, await db.Notifications.CountAsync(CancellationToken.None));
    }

    [Fact]
    public async Task CleanUpExpired_SentNotificationNotYetExpired_IsRetained()
    {
        await using var db = await CreateDbAsync();
        await SeedTerminalNotificationAsync(db.Context, PersistenceConstants.NotificationSent, expiresAt: FixedNow.AddDays(1));

        var dispatcher = CreateDispatcher(db.Context, new FakeTelegramClient());
        await dispatcher.CleanUpExpiredAsync(CancellationToken.None);

        Assert.Equal(1, await db.Notifications.CountAsync(CancellationToken.None));
    }

    [Fact]
    public async Task CleanUpExpired_PendingNotification_IsNotDeleted()
    {
        await using var db = await CreateDbAsync();
        // Pending notifications have a far-future expires_at
        await SeedJobAndNotificationAsync(db.Context, PersistenceConstants.JobFailed, errorMessage: "error");

        var dispatcher = CreateDispatcher(db.Context, new FakeTelegramClient());
        await dispatcher.CleanUpExpiredAsync(CancellationToken.None);

        Assert.Equal(1, await db.Notifications.CountAsync(CancellationToken.None));
    }

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------

    // TestDb keeps an open SqliteConnection so the :memory: database persists
    // across EF Core's internal open/close cycles for the test lifetime.
    private sealed class TestDb(SqliteConnection connection, ArchivistDbContext context) : IAsyncDisposable
    {
        public ArchivistDbContext Context { get; } = context;
        public DbSet<NotificationEntity> Notifications => Context.Notifications;
        public DbSet<JobEntity> Jobs => Context.Jobs;
        public DbSet<ArticleEntity> Articles => Context.Articles;

        public async ValueTask DisposeAsync()
        {
            await Context.DisposeAsync();
            await connection.DisposeAsync();
        }
    }

    private static async Task<TestDb> CreateDbAsync()
    {
        // CA2000: ownership of both connection and db is transferred to the TestDb
        // holder, which disposes both in DisposeAsync. The analyzer cannot follow this.
#pragma warning disable CA2000
        var connection = new SqliteConnection("Filename=:memory:");
        await connection.OpenAsync(CancellationToken.None);
        ArchivistDbContext? db = null;
        try
        {
            var options = new DbContextOptionsBuilder<ArchivistDbContext>()
                .UseSqlite(connection)
                .Options;
            db = new ArchivistDbContext(options);
            await db.Database.EnsureCreatedAsync(CancellationToken.None);
            return new TestDb(connection, db);
        }
        catch
        {
            if (db is not null) await db.DisposeAsync();
            await connection.DisposeAsync();
            throw;
        }
#pragma warning restore CA2000
    }

    private TelegramNotificationDispatcher CreateDispatcher(ArchivistDbContext db, FakeTelegramClient client)
    {
        var timeProvider = new FakeTimeProvider(FixedNow);
        var repo = new EfTelegramNotificationRepository(db);
        return new TelegramNotificationDispatcher(repo, client, timeProvider, NullLogger<TelegramNotificationDispatcher>.Instance);
    }

    private static async Task<(string JobId, string NotificationId)> SeedJobAndNotificationAsync(
        ArchivistDbContext db,
        string jobStatus,
        string? errorMessage)
    {
        const string userId = PersistenceConstants.PersonalUserId;
        const string articleId = "01ARTICLEID000000000000001";
        const string jobId = "01JOBID0000000000000000001";
        const string notificationId = "01NOTIFID00000000000000001";

        db.Users.Add(new UserEntity { Id = userId });
        db.Articles.Add(new ArticleEntity
        {
            Id = articleId,
            UserId = userId,
            OriginalUrl = "https://example.com",
            Status = jobStatus == PersistenceConstants.JobSucceeded ? PersistenceConstants.ArticleReady : PersistenceConstants.ArticleFailed,
            CreatedAt = FixedNow,
        });
        db.Jobs.Add(new JobEntity
        {
            Id = jobId,
            UserId = userId,
            ArticleId = articleId,
            Type = PersistenceConstants.ArticleProcessingJobType,
            Status = jobStatus,
            TelegramChatId = ChatId,
            TelegramMessageId = MessageId,
            TelegramUserId = 999,
            ErrorMessage = errorMessage,
            CreatedAt = FixedNow,
            CompletedAt = FixedNow,
            ExpiresAt = FixedNow.AddDays(14),
        });
        db.Notifications.Add(new NotificationEntity
        {
            Id = notificationId,
            JobId = jobId,
            Status = PersistenceConstants.NotificationPending,
            CreatedAt = FixedNow,
            ExpiresAt = FixedNow.AddDays(365),
        });
        await db.SaveChangesAsync(CancellationToken.None);

        return (jobId, notificationId);
    }

    private static async Task SeedTerminalNotificationAsync(
        ArchivistDbContext db,
        string notificationStatus,
        DateTimeOffset expiresAt)
    {
        const string userId = PersistenceConstants.PersonalUserId;
        const string articleId = "01ARTICLEID000000000000002";
        const string jobId = "01JOBID0000000000000000002";
        const string notificationId = "01NOTIFID00000000000000002";

        db.Users.Add(new UserEntity { Id = userId });
        db.Articles.Add(new ArticleEntity
        {
            Id = articleId,
            UserId = userId,
            OriginalUrl = "https://example.com",
            Status = PersistenceConstants.ArticleFailed,
            CreatedAt = FixedNow,
        });
        db.Jobs.Add(new JobEntity
        {
            Id = jobId,
            UserId = userId,
            ArticleId = articleId,
            Type = PersistenceConstants.ArticleProcessingJobType,
            Status = PersistenceConstants.JobFailed,
            TelegramChatId = ChatId,
            TelegramMessageId = MessageId,
            ErrorMessage = "some error",
            CreatedAt = FixedNow,
            CompletedAt = FixedNow,
            ExpiresAt = FixedNow.AddDays(14),
        });
        db.Notifications.Add(new NotificationEntity
        {
            Id = notificationId,
            JobId = jobId,
            Status = notificationStatus,
            CreatedAt = FixedNow,
            SentAt = notificationStatus == PersistenceConstants.NotificationSent ? FixedNow : null,
            ExpiresAt = expiresAt,
        });
        await db.SaveChangesAsync(CancellationToken.None);
    }

    // -------------------------------------------------------------------------
    // Fake Telegram client
    // -------------------------------------------------------------------------

    private sealed class FakeTelegramClient(bool failSend = false) : ITelegramClient
    {
        public List<SentReply> SentReplies { get; } = [];

        public Task SendReplyAsync(long chatId, long replyToMessageId, string text, CancellationToken cancellationToken)
        {
            if (failSend)
            {
                throw new InvalidOperationException("Fake Telegram delivery failure");
            }

            SentReplies.Add(new SentReply(chatId, replyToMessageId, text));
            return Task.CompletedTask;
        }
    }

    private sealed record SentReply(long ChatId, long ReplyToMessageId, string Text);
}