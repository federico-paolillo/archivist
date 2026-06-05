namespace Archivist.Gateway.Tests.Telegram;

using System.Diagnostics;

using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Telegram;

using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;

public sealed class TelegramWebhookHandlerTest
{
    private const string WebhookSecret = "test-webhook-secret";
    private const string MappedUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";
    private const long MappedTelegramUserId = 99999;
    private const long UnmappedTelegramUserId = 11111;
    private const long ChatId = 200;
    private const long MessageId = 300;

    [Fact]
    public async Task HandleAsync_AuthorizedMissingTextWithReplyTarget_ReturnsInvalidUrlAndSendsInvalidReply()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();
        var handler = CreateHandler(repo, client);

        var result = await handler.HandleAsync(
            new TelegramWebhookCommand(
                WebhookSecret,
                UpdateId: 1,
                SenderUserId: MappedTelegramUserId,
                ChatId: ChatId,
                MessageId: MessageId,
                MessageText: null),
            CancellationToken.None);

        Assert.Equal(TelegramWebhookOutcome.InvalidUrl, result.Outcome);
        Assert.False(repo.WasCalled);
        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(ChatId, reply.ChatId);
        Assert.Equal(MessageId, reply.ReplyToMessageId);
        Assert.Equal("Nope, you must send only an URL", reply.Text);
    }

    [Fact]
    public async Task HandleAsync_AuthorizedMissingReplyTarget_ReturnsNoMessageWithNoSideEffects()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();
        var handler = CreateHandler(repo, client);

        var result = await handler.HandleAsync(
            new TelegramWebhookCommand(
                WebhookSecret,
                UpdateId: 2,
                SenderUserId: MappedTelegramUserId,
                ChatId: null,
                MessageId: MessageId,
                MessageText: null),
            CancellationToken.None);

        Assert.Equal(TelegramWebhookOutcome.NoMessage, result.Outcome);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    [Fact]
    public async Task HandleAsync_ValidUrlInjectsCurrentTraceContextIntoIngestionCommand()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();
        var handler = CreateHandler(repo, client);

        using var root = new Activity("test-root");
        root.SetIdFormat(ActivityIdFormat.W3C);
        root.Start();

        await handler.HandleAsync(
            new TelegramWebhookCommand(
                WebhookSecret,
                UpdateId: 3,
                SenderUserId: MappedTelegramUserId,
                ChatId: ChatId,
                MessageId: MessageId,
                MessageText: "https://example.com/article"),
            CancellationToken.None);

        var command = Assert.Single(repo.RecordedCommands);

        Assert.NotNull(command.TraceParent);
        Assert.Contains(root.TraceId.ToHexString(), command.TraceParent, StringComparison.Ordinal);
        Assert.Equal(MappedUserId, command.UserId);
    }

    [Fact]
    public async Task HandleAsync_UnmappedSender_ReturnsUnauthorizedWithNoSideEffectsOrReply()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();
        var resolver = new FakeTelegramUserResolver();
        var handler = CreateHandler(repo, client, resolver);

        var result = await handler.HandleAsync(
            new TelegramWebhookCommand(
                WebhookSecret,
                UpdateId: 4,
                SenderUserId: UnmappedTelegramUserId,
                ChatId: ChatId,
                MessageId: MessageId,
                MessageText: "https://example.com/article"),
            CancellationToken.None);

        Assert.Equal(TelegramWebhookOutcome.Unauthorized, result.Outcome);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    private static TelegramWebhookHandler CreateHandler(
        FakeTelegramIngestionRepository repo,
        FakeTelegramClient client,
        FakeTelegramUserResolver? resolver = null)
    {
        var settings = Options.Create(new TelegramSettings
        {
            WebhookSecret = WebhookSecret,
            BotToken = "fake-token",
        });

        return new TelegramWebhookHandler(
            settings,
            resolver ?? new FakeTelegramUserResolver(),
            repo,
            client,
            NullLogger<TelegramWebhookHandler>.Instance);
    }

    private sealed class FakeTelegramClient : ITelegramClient
    {
        public List<SentReply> SentReplies { get; } = [];

        public Task SendReplyAsync(long chatId, long replyToMessageId, string text, CancellationToken cancellationToken)
        {
            SentReplies.Add(new SentReply(chatId, replyToMessageId, text));

            return Task.CompletedTask;
        }
    }

    private sealed record SentReply(long ChatId, long ReplyToMessageId, string Text);

    private sealed class FakeTelegramUserResolver : ITelegramUserResolver
    {
        public Task<string?> ResolveUserIdAsync(long telegramUserId, CancellationToken cancellationToken)
        {
            var userId = telegramUserId == MappedTelegramUserId ? MappedUserId : null;

            return Task.FromResult<string?>(userId);
        }
    }

    private sealed class FakeTelegramIngestionRepository : ITelegramIngestionRepository
    {
        public bool WasCalled { get; private set; }

        public List<RecordTelegramIngestionCommand> RecordedCommands { get; } = [];

        public Task<RecordTelegramIngestionResult> RecordValidUrlAsync(
            RecordTelegramIngestionCommand command,
            CancellationToken cancellationToken)
        {
            WasCalled = true;
            RecordedCommands.Add(command);

            return Task.FromResult(new RecordTelegramIngestionResult(
                Created: true,
                ArticleId: "01ASB2XFCZJY7WHZ2FNRTMQJC3",
                JobId: "01ASB2XFCZJY7WHZ2FNRTMQJC4"));
        }
    }
}