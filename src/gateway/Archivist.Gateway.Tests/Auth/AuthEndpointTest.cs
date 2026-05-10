using System.Net;
using System.Net.Http.Json;

using Archivist.Gateway.Application.Auth.Options;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.Extensions.Time.Testing;

using Xunit.Abstractions;

namespace Archivist.Gateway.Tests.Auth;

/// <summary>
/// Integration tests for <c>POST /login</c> and <c>POST /logout</c> endpoints.
/// </summary>
public sealed class AuthEndpointTest(ITestOutputHelper testOutputHelper) : IntegrationTest(testOutputHelper)
{
    private const string LoginPath = "/login";
    private const string LogoutPath = "/logout";
    private const string CookieName = "__Host-app-auth";
    private const int PasswordLength = 2048;
    private const string PersonalUserId = "01ASB2XFCZJY7WHZ2FNRTMQJCT";

    private static string ValidPassword() => new('a', PasswordLength);

    // -------------------------------------------------------------------------
    // POST /login success
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostLogin_ValidPassword_Returns204()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_SetsCookieWithCorrectName()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = response.Headers
            .GetValues("Set-Cookie")
            .FirstOrDefault(c => c.StartsWith(CookieName, StringComparison.Ordinal));

        Assert.NotNull(setCookie);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_CookieIsHttpOnly()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = GetAuthCookieHeader(response);

        Assert.NotNull(setCookie);
        Assert.Contains("HttpOnly", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_CookieIsSecure()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = GetAuthCookieHeader(response);

        Assert.NotNull(setCookie);
        Assert.Contains("Secure", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_CookieIsSameSiteStrict()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = GetAuthCookieHeader(response);

        Assert.NotNull(setCookie);
        Assert.Contains("SameSite=Strict", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_CookieHasPathSlash()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = GetAuthCookieHeader(response);

        Assert.NotNull(setCookie);
        Assert.Contains("Path=/", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_CookieHasNoDomain()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = GetAuthCookieHeader(response);

        Assert.NotNull(setCookie);
        Assert.DoesNotContain("Domain=", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_CookieHasNoExpires()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = GetAuthCookieHeader(response);

        Assert.NotNull(setCookie);
        Assert.DoesNotContain("Expires=", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task PostLogin_ValidPassword_CookieHasNoMaxAge()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin(ValidPassword());

        var setCookie = GetAuthCookieHeader(response);

        Assert.NotNull(setCookie);
        Assert.DoesNotContain("Max-Age=", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    // -------------------------------------------------------------------------
    // POST /login failures
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostLogin_WrongPassword_Returns401()
    {
        SetupSuccessEnvironment();

        var wrongPassword = new string('z', PasswordLength);
        var response = await SendLoginWithOrigin(wrongPassword);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_WrongPassword_DoesNotSetCookie()
    {
        SetupSuccessEnvironment();

        var wrongPassword = new string('z', PasswordLength);
        var response = await SendLoginWithOrigin(wrongPassword);

        var hasCookie = response.Headers.Contains("Set-Cookie") &&
            response.Headers.GetValues("Set-Cookie").Any(c => c.StartsWith(CookieName, StringComparison.Ordinal));

        Assert.False(hasCookie);
    }

    [Fact]
    public async Task PostLogin_OversizedPassword_Returns401()
    {
        SetupSuccessEnvironment();

        var oversized = new string('a', PasswordLength + 100);
        var response = await SendLoginWithOrigin(oversized);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_TooShortPassword_Returns401()
    {
        SetupSuccessEnvironment();

        var response = await SendLoginWithOrigin("short");

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_NullBody_Returns401()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        request.Headers.Add("Origin", "http://localhost");
        request.Content = JsonContent.Create<object?>(null);

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_MissingPasswordField_Returns401()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        request.Headers.Add("Origin", "http://localhost");
        request.Content = JsonContent.Create(new { });

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    // -------------------------------------------------------------------------
    // Session rotation on login
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostLogin_WithExistingValidSession_RotatesSession()
    {
        var (sessionStore, _, fakeTime) = SetupSuccessEnvironment();

        // Pre-seed an existing session.
        var existingSessionId = "pre-existing-session-id";
        var now = fakeTime.GetUtcNow();
        var existingEntry = new SessionEntry(PersonalUserId, now, now + TimeSpan.FromHours(24));
        await sessionStore.SetAsync(existingSessionId, existingEntry);

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        request.Headers.Add("Cookie", $"{CookieName}={existingSessionId}");
        request.Headers.Add("Origin", "http://localhost");
        request.Content = JsonContent.Create(new { password = ValidPassword() });

        await client.SendAsync(request);

        // Old session must be gone.
        var oldEntry = await sessionStore.GetAsync(existingSessionId);
        Assert.Null(oldEntry);
    }

    // -------------------------------------------------------------------------
    // Login throttling
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostLogin_ExceedPerIpLimit_Returns401WithoutPasswordVerification()
    {
        SetupSuccessEnvironment(
            configureServices: services =>
            {
                // Use a throttle that is already maxed out for all IPs.
                var throttle = new AlwaysThrottledLoginThrottle();
                services.AddSingleton<ILoginThrottle>(throttle);
            });

        // Even with the correct password, throttling should return 401.
        var response = await SendLoginWithOrigin(ValidPassword());

        Assert.Equal(HttpStatusCode.Unauthorized, response.StatusCode);
    }

    // -------------------------------------------------------------------------
    // POST /logout
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostLogout_WithValidSession_Returns204()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        request.Headers.Add("Cookie", $"{CookieName}=some-session");
        request.Headers.Add("Origin", "http://localhost");

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
    }

    [Fact]
    public async Task PostLogout_WithNoSession_Returns204()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        request.Headers.Add("Origin", "http://localhost");

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
    }

    [Fact]
    public async Task PostLogout_AlwaysClearsCookie()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        request.Headers.Add("Origin", "http://localhost");

        var response = await client.SendAsync(request);

        var setCookie = response.Headers
            .GetValues("Set-Cookie")
            .FirstOrDefault(c => c.StartsWith(CookieName, StringComparison.Ordinal));

        Assert.NotNull(setCookie);
        Assert.Contains("Max-Age=0", setCookie, StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public async Task PostLogout_WithValidSession_RemovesFromStore()
    {
        var (sessionStore, _, fakeTime) = SetupSuccessEnvironment();

        var sessionId = "session-to-logout";
        var now = fakeTime.GetUtcNow();
        await sessionStore.SetAsync(sessionId, new SessionEntry(PersonalUserId, now, now + TimeSpan.FromHours(24)));

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        request.Headers.Add("Cookie", $"{CookieName}={sessionId}");
        request.Headers.Add("Origin", "http://localhost");

        await client.SendAsync(request);

        var afterLogout = await sessionStore.GetAsync(sessionId);
        Assert.Null(afterLogout);
    }

    // -------------------------------------------------------------------------
    // Same-origin rejection
    // -------------------------------------------------------------------------

    [Fact]
    public async Task PostLogin_CrossOrigin_Returns403()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        request.Headers.Add("Origin", "https://evil.example.com");
        request.Content = JsonContent.Create(new { password = ValidPassword() });

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_NoOriginNoReferer_Returns403()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        request.Content = JsonContent.Create(new { password = ValidPassword() });

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Fact]
    public async Task PostLogout_CrossOrigin_Returns403()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        request.Headers.Add("Origin", "https://evil.example.com");

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------

    private (InMemorySessionStore SessionStore, FakePasswordStore PasswordStore, FakeTimeProvider Time) SetupSuccessEnvironment(
        Action<IServiceCollection>? configureServices = null)
    {
        var fakeTime = new FakeTimeProvider();
        var sessionStore = new InMemorySessionStore(fakeTime);
        var passwordHasher = new Argon2idPasswordHasher();
        var validHash = passwordHasher.Hash(ValidPassword());
        var passwordStore = new FakePasswordStore(validHash);

        PrepareEnvironment(services =>
        {
            services.AddSingleton<ISessionStore>(sessionStore);
            services.AddSingleton<IPasswordStore>(passwordStore);
            services.AddSingleton<TimeProvider>(fakeTime);

            configureServices?.Invoke(services);
        });

        return (sessionStore, passwordStore, fakeTime);
    }

    private async Task<HttpResponseMessage> SendLoginWithOrigin(string password)
    {
        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        request.Headers.Add("Origin", "http://localhost");
        request.Content = JsonContent.Create(new { password });

        return await client.SendAsync(request);
    }

    private static string? GetAuthCookieHeader(HttpResponseMessage response)
    {
        if (!response.Headers.Contains("Set-Cookie"))
        {
            return null;
        }

        return response.Headers
            .GetValues("Set-Cookie")
            .FirstOrDefault(c => c.StartsWith(CookieName, StringComparison.Ordinal));
    }

    // -------------------------------------------------------------------------
    // Fakes
    // -------------------------------------------------------------------------

    private sealed class FakePasswordStore(string? hash) : IPasswordStore
    {
        public Task<string?> GetPasswordHashAsync(CancellationToken ct = default) =>
            Task.FromResult(hash);
    }

    private sealed class AlwaysThrottledLoginThrottle : ILoginThrottle
    {
        public bool IsThrottled(string sourceIp) => true;
        public void RecordFailure(string sourceIp) { }
        public void RecordSuccess(string sourceIp) { }
    }
}