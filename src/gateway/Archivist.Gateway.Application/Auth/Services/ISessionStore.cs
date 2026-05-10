namespace Archivist.Gateway.Application.Auth.Services;

/// <summary>
/// Maps opaque session ids to their associated <see cref="SessionEntry"/> records.
/// </summary>
public interface ISessionStore
{
    /// <summary>
    /// Returns the session entry for the given session id, or <c>null</c> if not found.
    /// </summary>
    Task<SessionEntry?> GetAsync(string sessionId, CancellationToken ct = default);

    /// <summary>
    /// Stores or replaces the session entry for the given session id.
    /// </summary>
    Task SetAsync(string sessionId, SessionEntry entry, CancellationToken ct = default);

    /// <summary>
    /// Removes the session entry for the given session id. No-op if the key does not exist.
    /// </summary>
    Task RemoveAsync(string sessionId, CancellationToken ct = default);
}

/// <summary>
/// An authenticated session bound to a user with an absolute server-side expiry.
/// </summary>
public sealed record SessionEntry(
    string UserId,
    DateTimeOffset CreatedAt,
    DateTimeOffset AbsoluteExpiresAt);