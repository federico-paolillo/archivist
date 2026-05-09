namespace Archivist.Gateway.Api.Telegram;

using Archivist.Gateway.Api.Telegram.Models;
using Archivist.Gateway.Application.Telegram;

using Microsoft.AspNetCore.Http.HttpResults;

/// <summary>
/// Static handler methods for Telegram webhook API routes.
/// </summary>
internal static class Handlers
{
    private const string SecretTokenHeader = "X-Telegram-Bot-Api-Secret-Token";

    /// <summary>
    /// Handles POST /telegram/webhook. Validates the webhook secret, authorizes the sender,
    /// validates URL-only text, and enqueues valid URLs.
    /// </summary>
    public static async Task<Ok> PostWebhook(
        TelegramUpdateDto update,
        HttpRequest request,
        TelegramWebhookHandler handler,
        CancellationToken cancellationToken)
    {
        ArgumentNullException.ThrowIfNull(update);
        ArgumentNullException.ThrowIfNull(request);
        ArgumentNullException.ThrowIfNull(handler);

        var secret = request.Headers[SecretTokenHeader].FirstOrDefault();

        var command = new TelegramWebhookCommand(
            WebhookSecret: secret,
            UpdateId: update.UpdateId,
            SenderUserId: update.Message?.From?.Id,
            ChatId: update.Message?.Chat?.Id,
            MessageId: update.Message?.MessageId,
            MessageText: update.Message?.Text);

        await handler.HandleAsync(command, cancellationToken).ConfigureAwait(false);

        return TypedResults.Ok();
    }
}