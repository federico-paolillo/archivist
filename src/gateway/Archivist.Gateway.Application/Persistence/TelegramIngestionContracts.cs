namespace Archivist.Gateway.Application.Persistence;

using System.Diagnostics.CodeAnalysis;

/// <summary>
/// Captures the metadata required to persist an accepted Telegram URL.
/// </summary>
[SuppressMessage("Design", "CA1054:URI-like parameters should not be strings", Justification = "Persistence contract stores canonical SQLite URL text.")]
[SuppressMessage("Design", "CA1056:URI-like properties should not be strings", Justification = "Persistence contract stores canonical SQLite URL text.")]
public sealed record RecordTelegramIngestionCommand(
    long TelegramUpdateId,
    long TelegramChatId,
    long TelegramMessageId,
    long TelegramUserId,
    string UserId,
    string OriginalUrl,
    string? TraceParent = null,
    string? TraceState = null);

/// <summary>
/// Describes the outcome of recording a Telegram URL ingestion.
/// </summary>
public sealed record RecordTelegramIngestionResult(
    bool Created,
    string ArticleId,
    string JobId);

/// <summary>
/// Persists accepted Telegram URL ingestion records atomically.
/// </summary>
public interface ITelegramIngestionRepository
{
    /// <summary>
    /// Creates the article and queued job for the resolved user unless the update was already seen.
    /// </summary>
    Task<RecordTelegramIngestionResult> RecordValidUrlAsync(
        RecordTelegramIngestionCommand command,
        CancellationToken cancellationToken);
}