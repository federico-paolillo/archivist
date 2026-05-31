namespace Archivist.Gateway.Tests.Api;

using System.Net;
using System.Net.Http.Json;
using System.Text.Json.Serialization;

using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Telegram;

using Microsoft.Extensions.DependencyInjection;

using Xunit.Abstractions;

public sealed class TelegramWebhookEndpointTest(ITestOutputHelper testOutputHelper) : IntegrationTest(testOutputHelper)
{
    private const string WebhookPath = "/telegram/webhook";
    private const string SecretHeader = "X-Telegram-Bot-Api-Secret-Token";
    private const string ValidSecret = "test-webhook-secret";
    private const long AllowedUserId = 99999;
    private const long AnotherUserId = 11111;
    private const long ChatId = 200;
    private const long MessageId = 300;

    // -------------------------------------------------------------------------
    // Bad secret
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostWebhook_BadSecret_Returns200WithNoSideEffects()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(1, AllowedUserId, ChatId, MessageId, "https://example.com");
        using var request = new HttpRequestMessage(HttpMethod.Post, WebhookPath)
        {
            Content = JsonContent.Create(update),
        };
        request.Headers.Add(SecretHeader, "wrong-secret");

        var response = await http.SendAsync(request);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    [Fact]
    public async Task PostWebhook_MissingSecret_Returns200WithNoSideEffects()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(1, AllowedUserId, ChatId, MessageId, "https://example.com");
        var response = await http.PostAsJsonAsync(WebhookPath, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    // -------------------------------------------------------------------------
    // Unauthorized sender
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostWebhook_UnauthorizedSender_Returns200WithNoSideEffects()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(2, AnotherUserId, ChatId, MessageId, "https://example.com");
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    // -------------------------------------------------------------------------
    // Invalid URL (authorized sender)
    // -------------------------------------------------------------------------

    [Theory]
    [InlineData("read this please https://example.com")]
    [InlineData("ftp://example.com")]
    [InlineData("not a url at all")]
    [InlineData("https://example.com extra text")]
    public async Task PostWebhook_InvalidUrl_Returns200AndSendsInvalidReply(string text)
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(3, AllowedUserId, ChatId, MessageId, text);
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(ChatId, reply.ChatId);
        Assert.Equal(MessageId, reply.ReplyToMessageId);
        Assert.Equal("Nope, you must send only an URL", reply.Text);
    }

    [Fact]
    public async Task PostWebhook_AuthorizedMediaOnlyMessage_Returns200AndSendsInvalidReply()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildMediaOnlyUpdate(4, AllowedUserId, ChatId, MessageId);
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(ChatId, reply.ChatId);
        Assert.Equal(MessageId, reply.ReplyToMessageId);
        Assert.Equal("Nope, you must send only an URL", reply.Text);
    }

    [Fact]
    public async Task PostWebhook_AuthorizedCaptionMessage_Returns200AndSendsInvalidReply()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildCaptionUpdate(5, AllowedUserId, ChatId, MessageId, "https://example.com/article");
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(ChatId, reply.ChatId);
        Assert.Equal(MessageId, reply.ReplyToMessageId);
        Assert.Equal("Nope, you must send only an URL", reply.Text);
    }

    // -------------------------------------------------------------------------
    // Valid URL
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostWebhook_ValidUrl_Returns200AndSendsAcknowledgement()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(10, AllowedUserId, ChatId, MessageId, "https://example.com/article");
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.True(repo.WasCalled);

        var cmd = Assert.Single(repo.RecordedCommands);
        Assert.Equal(10, cmd.TelegramUpdateId);
        Assert.Equal(ChatId, cmd.TelegramChatId);
        Assert.Equal(MessageId, cmd.TelegramMessageId);
        Assert.Equal(AllowedUserId, cmd.TelegramUserId);
        Assert.Equal("https://example.com/article", cmd.OriginalUrl);

        var reply = Assert.Single(client.SentReplies);
        Assert.Equal(ChatId, reply.ChatId);
        Assert.Equal(MessageId, reply.ReplyToMessageId);
        Assert.Equal("Ok, I will have a look", reply.Text);
    }

    [Fact]
    public async Task PostWebhook_ValidUrl_SenderUserIdIsDistinctFromChatId()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        const long distinctChatId = 555555;
        const long senderUserId = AllowedUserId;

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(20, senderUserId, distinctChatId, MessageId, "https://example.com/article");
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.True(repo.WasCalled);

        var cmd = Assert.Single(repo.RecordedCommands);
        Assert.Equal(senderUserId, cmd.TelegramUserId);
        Assert.Equal(distinctChatId, cmd.TelegramChatId);
        Assert.NotEqual(cmd.TelegramUserId, cmd.TelegramChatId);
    }

    // -------------------------------------------------------------------------
    // Duplicate update
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostWebhook_DuplicateUpdate_Returns200AndNoReply()
    {
        var repo = new FakeTelegramIngestionRepository(isDuplicate: true);
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(50, AllowedUserId, ChatId, MessageId, "https://example.com/article");
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.True(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    // -------------------------------------------------------------------------
    // Acknowledgement failure does not roll back ingestion
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostWebhook_AcknowledgementFails_Returns200AndJobPersists()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient(failSend: true);

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(60, AllowedUserId, ChatId, MessageId, "https://example.com/article");
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.True(repo.WasCalled);
        Assert.Single(repo.RecordedCommands);
    }

    // -------------------------------------------------------------------------
    // No message / non-text update
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostWebhook_NoMessage_Returns200WithNoSideEffects()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = new UpdatePayload { UpdateId = 70 };
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    [Fact]
    public async Task PostWebhook_AuthorizedMessageWithoutReplyTarget_Returns200WithNoSideEffects()
    {
        var repo = new FakeTelegramIngestionRepository();
        var client = new FakeTelegramClient();

        PrepareWebhookEnvironment(repo, client);

        using var http = CreateHttpClient();
        var update = BuildTextUpdate(71, AllowedUserId, chatId: 0, messageId: null, "https://example.com/article");
        update.Message!.Chat = null;
        var response = await SendWithSecret(http, update);

        Assert.Equal(HttpStatusCode.OK, response.StatusCode);
        Assert.False(repo.WasCalled);
        Assert.Empty(client.SentReplies);
    }

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------

    private void PrepareWebhookEnvironment(
        FakeTelegramIngestionRepository repo,
        FakeTelegramClient client)
    {
        var sqlitePath = Path.Combine(Path.GetTempPath(), $"{Guid.NewGuid():N}.db");

        PrepareEnvironment(
            Environments.Development,
            configureTestServices: services =>
            {
                services.AddSingleton<ITelegramIngestionRepository>(repo);
                services.AddSingleton<ITelegramClient>(client);
            },
            configureConfiguration: cfg =>
                cfg.AddInMemoryCollection(new Dictionary<string, string?>
                {
                    ["SQLITE_PATH"] = sqlitePath,
                    ["Telegram:WebhookSecret"] = ValidSecret,
                    ["Telegram:AllowedUserId"] = AllowedUserId.ToString(),
                    ["Telegram:BotToken"] = "fake-token",
                }));
    }

    private static async Task<HttpResponseMessage> SendWithSecret(HttpClient http, object payload)
    {
        using var request = new HttpRequestMessage(HttpMethod.Post, WebhookPath)
        {
            Content = JsonContent.Create(payload),
        };
        request.Headers.Add(SecretHeader, ValidSecret);

        return await http.SendAsync(request).ConfigureAwait(false);
    }

    private static UpdatePayload BuildTextUpdate(
        long updateId,
        long fromUserId,
        long chatId,
        long? messageId,
        string text)
    {
        return new UpdatePayload
        {
            UpdateId = updateId,
            Message = new MessagePayload
            {
                MessageId = messageId,
                From = new UserPayload { Id = fromUserId },
                Chat = new ChatPayload { Id = chatId },
                Text = text,
            },
        };
    }

    private static UpdatePayload BuildMediaOnlyUpdate(
        long updateId,
        long fromUserId,
        long chatId,
        long messageId)
    {
        return new UpdatePayload
        {
            UpdateId = updateId,
            Message = new MessagePayload
            {
                MessageId = messageId,
                From = new UserPayload { Id = fromUserId },
                Chat = new ChatPayload { Id = chatId },
                Photo = [new PhotoPayload { FileId = "photo-file-id" }],
            },
        };
    }

    private static UpdatePayload BuildCaptionUpdate(
        long updateId,
        long fromUserId,
        long chatId,
        long messageId,
        string caption)
    {
        return new UpdatePayload
        {
            UpdateId = updateId,
            Message = new MessagePayload
            {
                MessageId = messageId,
                From = new UserPayload { Id = fromUserId },
                Chat = new ChatPayload { Id = chatId },
                Caption = caption,
                Photo = [new PhotoPayload { FileId = "photo-file-id" }],
            },
        };
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
                throw new InvalidOperationException("Fake Telegram send failure");
            }

            SentReplies.Add(new SentReply(chatId, replyToMessageId, text));

            return Task.CompletedTask;
        }
    }

    private sealed record SentReply(long ChatId, long ReplyToMessageId, string Text);

    // -------------------------------------------------------------------------
    // Fake ingestion repository
    // -------------------------------------------------------------------------

    private sealed class FakeTelegramIngestionRepository(bool isDuplicate = false) : ITelegramIngestionRepository
    {
        public bool WasCalled { get; private set; }

        public List<RecordTelegramIngestionCommand> RecordedCommands { get; } = [];

        public Task<RecordTelegramIngestionResult> RecordValidUrlAsync(
            RecordTelegramIngestionCommand command,
            CancellationToken cancellationToken)
        {
            WasCalled = true;
            RecordedCommands.Add(command);

            var result = isDuplicate
                ? new RecordTelegramIngestionResult(false, "01ASB2XFCZJY7WHZ2FNRTMQJC1", "01ASB2XFCZJY7WHZ2FNRTMQJC2")
                : new RecordTelegramIngestionResult(true, "01ASB2XFCZJY7WHZ2FNRTMQJC3", "01ASB2XFCZJY7WHZ2FNRTMQJC4");

            return Task.FromResult(result);
        }
    }

    // -------------------------------------------------------------------------
    // JSON payloads for update construction (mirrors TelegramUpdateDto shape)
    // -------------------------------------------------------------------------

    private sealed class UpdatePayload
    {
        [JsonPropertyName("update_id")]
        public long UpdateId { get; set; }

        [JsonPropertyName("message")]
        public MessagePayload? Message { get; set; }
    }

    private sealed class MessagePayload
    {
        [JsonPropertyName("message_id")]
        public long? MessageId { get; set; }

        [JsonPropertyName("from")]
        public UserPayload? From { get; set; }

        [JsonPropertyName("chat")]
        public ChatPayload? Chat { get; set; }

        [JsonPropertyName("text")]
        public string? Text { get; set; }

        [JsonPropertyName("caption")]
        public string? Caption { get; set; }

        [JsonPropertyName("photo")]
        public PhotoPayload[]? Photo { get; set; }
    }

    private sealed class PhotoPayload
    {
        [JsonPropertyName("file_id")]
        public string FileId { get; set; } = string.Empty;
    }

    private sealed class UserPayload
    {
        [JsonPropertyName("id")]
        public long Id { get; set; }
    }

    private sealed class ChatPayload
    {
        [JsonPropertyName("id")]
        public long Id { get; set; }
    }
}