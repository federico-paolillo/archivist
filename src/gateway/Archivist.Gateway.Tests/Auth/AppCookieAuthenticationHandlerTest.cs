using System.Net;

using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Routing;
using Microsoft.Extensions.Time.Testing;

using Xunit.Abstractions;

namespace Archivist.Gateway.Tests.Auth;

/// <summary>
/// Integration tests for <see cref="AppCookieAuthenticationHandler"/> using the real DI pipeline
/// via the <c>GET /auth/session</c> endpoint as the authentication probe.
/// </summary>
public sealed class AppCookieAuthenticationHandlerTest(ITestOutputHelper testOutputHelper) : IntegrationTest(testOutputHelper)
{
    private const string CookieName = "__Host-app-auth";
    private const string PersonalUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";

    [Fact]
    public async Task GetSession_NoCookie_Returns401()
    {
        var sessionStore = new InMemorySessionStore(new FakeTimeProvider());
        PrepareEnvironment(services =>
        {
            services.AddSingleton<ISessionStore>(sessionStore);
        });

        using var client = CreateHttpClient();
        var response = await client.GetAsync("/auth/session");

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public void ProtectedUiApiProbe_IsMappedWithAuthorizationMetadata()
    {
        var sessionStore = new InMemorySessionStore(new FakeTimeProvider());
        PrepareEnvironment(services =>
        {
            services.AddSingleton<ISessionStore>(sessionStore);
        });

        var dataSource = GetRequiredService<EndpointDataSource>();
        var endpoint = dataSource.Endpoints
            .OfType<RouteEndpoint>()
            .Single(e => e.RoutePattern.RawText == "/auth/session");

        Assert.Contains(endpoint.Metadata, metadata => metadata is IAuthorizeData);
    }

    [Fact]
    public async Task GetSession_ValidSession_Returns204()
    {
        var fakeTime = new FakeTimeProvider();
        var sessionStore = new InMemorySessionStore(fakeTime);

        var now = fakeTime.GetUtcNow();
        var entry = new SessionEntry(PersonalUserId, now, now + TimeSpan.FromHours(24));
        await sessionStore.SetAsync("valid-session-id", entry);

        PrepareEnvironment(services =>
        {
            services.AddSingleton<ISessionStore>(sessionStore);
            services.AddSingleton<TimeProvider>(fakeTime);
        });

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Get, "/auth/session");
        request.Headers.Add("Cookie", $"{CookieName}=valid-session-id");

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
    }

    [Fact]
    public async Task GetSession_UnknownSessionId_Returns401()
    {
        var sessionStore = new InMemorySessionStore(new FakeTimeProvider());
        PrepareEnvironment(services =>
        {
            services.AddSingleton<ISessionStore>(sessionStore);
        });

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Get, "/auth/session");
        request.Headers.Add("Cookie", $"{CookieName}=nonexistent-session-id");

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task GetSession_ExpiredSession_Returns401()
    {
        var fakeTime = new FakeTimeProvider();
        var sessionStore = new InMemorySessionStore(fakeTime);

        var now = fakeTime.GetUtcNow();
        var entry = new SessionEntry(PersonalUserId, now, now + TimeSpan.FromMinutes(30));
        await sessionStore.SetAsync("expired-session-id", entry);

        // Advance time past expiry.
        fakeTime.Advance(TimeSpan.FromHours(1));

        PrepareEnvironment(services =>
        {
            services.AddSingleton<ISessionStore>(sessionStore);
            services.AddSingleton<TimeProvider>(fakeTime);
        });

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Get, "/auth/session");
        request.Headers.Add("Cookie", $"{CookieName}=expired-session-id");

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task GetSession_ExpiredSession_RemovesEntryFromStore()
    {
        var fakeTime = new FakeTimeProvider();
        var sessionStore = new InMemorySessionStore(fakeTime);

        var now = fakeTime.GetUtcNow();
        var entry = new SessionEntry(PersonalUserId, now, now + TimeSpan.FromMinutes(30));
        await sessionStore.SetAsync("expired-and-removed-id", entry);

        fakeTime.Advance(TimeSpan.FromHours(1));

        PrepareEnvironment(services =>
        {
            services.AddSingleton<ISessionStore>(sessionStore);
            services.AddSingleton<TimeProvider>(fakeTime);
        });

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Get, "/auth/session");
        request.Headers.Add("Cookie", $"{CookieName}=expired-and-removed-id");

        await client.SendAsync(request);

        // After authentication attempt, store should no longer contain the expired entry.
        var afterLookup = await sessionStore.GetAsync("expired-and-removed-id");
        Assert.Null(afterLookup);
    }
}