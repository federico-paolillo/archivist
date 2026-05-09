namespace Archivist.Gateway.Application.Telegram;

/// <summary>
/// Sends messages through the Telegram Bot API.
/// </summary>
public interface ITelegramClient
{
    /// <summary>
    /// Sends a text reply to an existing Telegram message.
    /// </summary>
    Task SendReplyAsync(long chatId, long replyToMessageId, string text, CancellationToken cancellationToken);
}