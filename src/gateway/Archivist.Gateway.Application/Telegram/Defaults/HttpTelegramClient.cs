namespace Archivist.Gateway.Application.Telegram.Defaults;

using System.Net.Http.Json;
using System.Text.Json.Serialization;

using Microsoft.Extensions.Options;

/// <summary>
/// Sends Telegram Bot API messages over HTTPS.
/// </summary>
public sealed class HttpTelegramClient(
    HttpClient httpClient,
    IOptions<TelegramSettings> settings) : ITelegramClient
{
    /// <inheritdoc />
    public async Task SendReplyAsync(long chatId, long replyToMessageId, string text, CancellationToken cancellationToken)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(text);

        var botToken = settings.Value.BotToken;
        ArgumentException.ThrowIfNullOrWhiteSpace(botToken);

        var payload = new SendMessagePayload(chatId, text, new ReplyParameters(replyToMessageId));
        var url = $"https://api.telegram.org/bot{botToken}/sendMessage";

        var response = await httpClient
            .PostAsJsonAsync(url, payload, cancellationToken)
            .ConfigureAwait(false);

        response.EnsureSuccessStatusCode();
    }

    private sealed record SendMessagePayload(
        [property: JsonPropertyName("chat_id")] long ChatId,
        [property: JsonPropertyName("text")] string Text,
        [property: JsonPropertyName("reply_parameters")] ReplyParameters ReplyParameters);

    private sealed record ReplyParameters(
        [property: JsonPropertyName("message_id")] long MessageId);
}