namespace Archivist.Gateway.Application.Telegram;

/// <summary>
/// Describes the possible outcomes of processing a Telegram webhook update.
/// </summary>
public enum TelegramWebhookOutcome
{
    /// <summary>
    /// The webhook secret was invalid; the update was rejected without processing.
    /// </summary>
    BadSecret,

    /// <summary>
    /// The sender is not the configured allowed user; no side effects were produced.
    /// </summary>
    Unauthorized,

    /// <summary>
    /// The update had no processable text message; no side effects were produced.
    /// </summary>
    NoMessage,

    /// <summary>
    /// The message text was not a valid absolute http/https URL; the invalid reply was sent.
    /// </summary>
    InvalidUrl,

    /// <summary>
    /// A duplicate update_id was received; no new records were created.
    /// </summary>
    Duplicate,

    /// <summary>
    /// The URL was enqueued and the acknowledgement reply was sent.
    /// </summary>
    Queued,

    /// <summary>
    /// The URL was enqueued but the acknowledgement reply failed; the job persists.
    /// </summary>
    QueuedReplyFailed,
}

/// <summary>
/// Describes the result of processing a Telegram webhook update.
/// </summary>
public sealed record TelegramWebhookResult(TelegramWebhookOutcome Outcome);