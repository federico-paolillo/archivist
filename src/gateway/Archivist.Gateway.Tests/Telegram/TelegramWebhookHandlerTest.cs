namespace Archivist.Gateway.Tests.Telegram;

using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Telegram;

using Microsoft.Extensions.Logging.Abstractions;
using Microsoft.Extensions.Options;

public sealed class TelegramWebhookHandlerTest
{
    private const string WebhookSecret = "test-webhook-secret";
    private const long AllowedUserId = 99999;
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
                SenderUserId: AllowedUserId,
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
                SenderUserId: AllowedUserId,
                ChatId: null,
                MessageId: MessageId,
                MessageText: null),
            CancellationToken.None);

        Assert.Equal(TelegramWebhookOutcome.NoMessage, result.Outcome);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    private static TelegramWebhookHandler CreateHandler(
        FakeTelegramIngestionRepository repo,
        FakeTelegramClient client)
    {
        var settings = Options.Create(new TelegramSettings
        {
            WebhookSecret = WebhookSecret,
            AllowedUserId = AllowedUserId,
            BotToken = "fake-token",
        });

        return new TelegramWebhookHandler(
            settings,
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

    private sealed class FakeTelegramIngestionRepository : ITelegramIngestionRepository
    {
        public bool WasCalled { get; private set; }

        public Task<RecordTelegramIngestionResult> RecordValidUrlAsync(
            RecordTelegramIngestionCommand command,
            CancellationToken cancellationToken)
        {
            WasCalled = true;

            return Task.FromResult(new RecordTelegramIngestionResult(
                Created: true,
                ArticleId: "01ASB2XFCZJY7WHZ2FNRTMQJC3",
                JobId: "01ASB2XFCZJY7WHZ2FNRTMQJC4"));
        }
    }
}