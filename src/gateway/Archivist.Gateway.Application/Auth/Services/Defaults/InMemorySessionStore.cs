using System.Collections.Concurrent;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// V0 in-memory implementation of <see cref="ISessionStore"/> backed by a <see cref="ConcurrentDictionary{TKey,TValue}"/>.
/// All sessions are lost on gateway restart by design.
/// Expired entries are removed on lookup.
/// </summary>
public sealed class InMemorySessionStore(TimeProvider timeProvider) : ISessionStore
{
    private readonly ConcurrentDictionary<string, SessionEntry> _store = new(StringComparer.Ordinal);

    public Task<SessionEntry?> GetAsync(string sessionId, CancellationToken ct = default)
    {
        ArgumentNullException.ThrowIfNull(sessionId);

        if (!_store.TryGetValue(sessionId, out var entry))
        {
            return Task.FromResult<SessionEntry?>(null);
        }

        if (timeProvider.GetUtcNow() >= entry.AbsoluteExpiresAt)
        {
            // Remove expired entry eagerly on lookup.
            _store.TryRemove(sessionId, out _);
            return Task.FromResult<SessionEntry?>(null);
        }

        return Task.FromResult<SessionEntry?>(entry);
    }

    public Task SetAsync(string sessionId, SessionEntry entry, CancellationToken ct = default)
    {
        ArgumentNullException.ThrowIfNull(sessionId);
        ArgumentNullException.ThrowIfNull(entry);

        _store[sessionId] = entry;
        return Task.CompletedTask;
    }

    public Task RemoveAsync(string sessionId, CancellationToken ct = default)
    {
        ArgumentNullException.ThrowIfNull(sessionId);

        _store.TryRemove(sessionId, out _);
        return Task.CompletedTask;
    }
}