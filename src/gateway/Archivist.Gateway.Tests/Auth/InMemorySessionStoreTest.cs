using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.Extensions.Time.Testing;

namespace Archivist.Gateway.Tests.Auth;

/// <summary>
/// Unit tests for <see cref="InMemorySessionStore"/>.
/// </summary>
public sealed class InMemorySessionStoreTest
{
    private static SessionEntry MakeEntry(DateTimeOffset createdAt, TimeSpan lifetime, string userId = "user-1") =>
        new(userId, createdAt, createdAt + lifetime);

    [Fact]
    public async Task GetAsync_UnknownSession_ReturnsNull()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        var result = await store.GetAsync("unknown-session-id");

        Assert.Null(result);
    }

    [Fact]
    public async Task SetAndGetAsync_ValidSession_ReturnsEntry()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        var now = timeProvider.GetUtcNow();
        var entry = MakeEntry(now, TimeSpan.FromHours(24));

        await store.SetAsync("session-1", entry);

        var result = await store.GetAsync("session-1");

        Assert.NotNull(result);
        Assert.Equal("user-1", result.UserId);
    }

    [Fact]
    public async Task GetAsync_ExpiredSession_ReturnsNull()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        var now = timeProvider.GetUtcNow();
        var entry = MakeEntry(now, TimeSpan.FromHours(1));

        await store.SetAsync("expired-session", entry);

        // Advance clock past expiry.
        timeProvider.Advance(TimeSpan.FromHours(2));

        var result = await store.GetAsync("expired-session");

        Assert.Null(result);
    }

    [Fact]
    public async Task GetAsync_ExpiredSession_RemovesEntryFromStore()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        var now = timeProvider.GetUtcNow();
        var entry = MakeEntry(now, TimeSpan.FromMinutes(30));

        await store.SetAsync("stale-session", entry);

        // Expire the session.
        timeProvider.Advance(TimeSpan.FromHours(1));

        // First lookup should remove the entry.
        await store.GetAsync("stale-session");

        // Second lookup should still return null (not re-inserted).
        var result = await store.GetAsync("stale-session");
        Assert.Null(result);
    }

    [Fact]
    public async Task RemoveAsync_ExistingSession_RemovesEntry()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        var now = timeProvider.GetUtcNow();
        var entry = MakeEntry(now, TimeSpan.FromHours(24));

        await store.SetAsync("session-to-remove", entry);
        await store.RemoveAsync("session-to-remove");

        var result = await store.GetAsync("session-to-remove");
        Assert.Null(result);
    }

    [Fact]
    public async Task RemoveAsync_MissingSession_IsNoOp()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        // Should not throw.
        var ex = await Record.ExceptionAsync(() => store.RemoveAsync("nonexistent"));
        Assert.Null(ex);
    }

    [Fact]
    public async Task SetAsync_OverwritesExistingEntry()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        var now = timeProvider.GetUtcNow();
        var original = new SessionEntry("user-original", now, now + TimeSpan.FromHours(24));
        var updated = new SessionEntry("user-updated", now, now + TimeSpan.FromHours(24));

        await store.SetAsync("session-1", original);
        await store.SetAsync("session-1", updated);

        var result = await store.GetAsync("session-1");

        Assert.NotNull(result);
        Assert.Equal("user-updated", result.UserId);
    }

    [Fact]
    public async Task GetAsync_SessionExpiresAtBoundary_ReturnsNull()
    {
        var timeProvider = new FakeTimeProvider();
        var store = new InMemorySessionStore(timeProvider);

        var now = timeProvider.GetUtcNow();
        var entry = MakeEntry(now, TimeSpan.FromHours(1));

        await store.SetAsync("boundary-session", entry);

        // Advance exactly to expiry boundary.
        timeProvider.Advance(TimeSpan.FromHours(1));

        var result = await store.GetAsync("boundary-session");

        Assert.Null(result);
    }
}