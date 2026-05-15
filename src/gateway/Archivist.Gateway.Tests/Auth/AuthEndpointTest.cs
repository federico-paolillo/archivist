using System.Net;
using System.Net.Http.Json;

using Archivist.Gateway.Application.Auth.Options;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.Extensions.Hosting;
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
    private const string PublicHost = "localhost";
    private const string PublicOrigin = "https://localhost";

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
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);
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
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);
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
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);
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
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
    }

    [Fact]
    public async Task PostLogout_WithNoSession_Returns204()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.NoContent, response.StatusCode);
    }

    [Fact]
    public async Task PostLogout_AlwaysClearsCookie()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);

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
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);

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
        AddTrustedForwardedHeaders(request);
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
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", "https://evil.example.com");

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_EffectiveHttp_Returns403AndDoesNotSetCookie()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        AddTrustedForwardedHeaders(request, scheme: "http");
        request.Headers.Add("Origin", "http://localhost");
        request.Content = JsonContent.Create(new { password = ValidPassword() });

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
        Assert.Null(GetAuthCookieHeader(response));
    }

    [Theory]
    [InlineData("http://localhost", null)]
    [InlineData("https://example.net", null)]
    [InlineData("https://localhost:9443", "localhost:8443")]
    public async Task PostLogin_OriginMismatch_Returns403(string origin, string? host)
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", origin);
        request.Content = JsonContent.Create(new { password = ValidPassword() });

        if (host is not null)
        {
            request.Headers.Host = host;
            request.Headers.Remove("X-Forwarded-Host");
        }

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Theory]
    [InlineData("http://localhost", null)]
    [InlineData("https://example.net", null)]
    [InlineData("https://localhost:9443", "localhost:8443")]
    public async Task PostLogout_OriginMismatch_Returns403(string origin, string? host)
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LogoutPath);
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", origin);

        if (host is not null)
        {
            request.Headers.Host = host;
            request.Headers.Remove("X-Forwarded-Host");
        }

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Fact]
    public async Task PostLogin_DisallowedForwardedHost_Returns403()
    {
        SetupSuccessEnvironment(publicHosts: "allowed.example");

        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        AddTrustedForwardedHeaders(request, host: "blocked.example");
        request.Headers.Add("Origin", "https://blocked.example");
        request.Content = JsonContent.Create(new { password = ValidPassword() });

        var response = await client.SendAsync(request);

        Assert.Equal(HttpStatusCode.Forbidden, response.StatusCode);
    }

    [Fact]
    public void MissingGatewayPublicHosts_FailsFastOutsideDevelopment()
    {
        PrepareEnvironment(Environments.Production);

        var exception = Assert.Throws<InvalidOperationException>(() => CreateHttpClient());

        Assert.Contains("GATEWAY_PUBLIC_HOSTS", exception.Message, StringComparison.Ordinal);
    }

    [Fact]
    public async Task ApiPrefixedAuthRoutes_AreNotMappedByGateway()
    {
        SetupSuccessEnvironment();

        using var client = CreateHttpClient();

        using var login = new HttpRequestMessage(HttpMethod.Post, "/api/login");
        AddTrustedForwardedHeaders(login);
        login.Headers.Add("Origin", PublicOrigin);
        login.Content = JsonContent.Create(new { password = ValidPassword() });

        using var logout = new HttpRequestMessage(HttpMethod.Post, "/api/logout");
        AddTrustedForwardedHeaders(logout);
        logout.Headers.Add("Origin", PublicOrigin);

        using var session = new HttpRequestMessage(HttpMethod.Get, "/api/auth/session");
        AddTrustedForwardedHeaders(session);

        var loginResponse = await client.SendAsync(login);
        var logoutResponse = await client.SendAsync(logout);
        var sessionResponse = await client.SendAsync(session);

        Assert.Equal(HttpStatusCode.NotFound, loginResponse.StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, logoutResponse.StatusCode);
        Assert.Equal(HttpStatusCode.NotFound, sessionResponse.StatusCode);
    }

    // -------------------------------------------------------------------------
    // Helpers
    // -------------------------------------------------------------------------

    private (InMemorySessionStore SessionStore, FakePasswordStore PasswordStore, FakeTimeProvider Time) SetupSuccessEnvironment(
        Action<IServiceCollection>? configureServices = null,
        string publicHosts = PublicHost)
    {
        var fakeTime = new FakeTimeProvider();
        var sessionStore = new InMemorySessionStore(fakeTime);
        var passwordHasher = new Argon2idPasswordHasher();
        var validHash = passwordHasher.Hash(ValidPassword());
        var passwordStore = new FakePasswordStore(validHash);

        PrepareEnvironment(
            Environments.Development,
            configureTestServices: services =>
            {
                services.AddSingleton<ISessionStore>(sessionStore);
                services.AddSingleton<IPasswordStore>(passwordStore);
                services.AddSingleton<TimeProvider>(fakeTime);

                configureServices?.Invoke(services);
            },
            configureConfiguration: cfg =>
                cfg.AddInMemoryCollection(new Dictionary<string, string?>
                {
                    ["GATEWAY_PUBLIC_HOSTS"] = publicHosts,
                }));

        return (sessionStore, passwordStore, fakeTime);
    }

    private async Task<HttpResponseMessage> SendLoginWithOrigin(string password)
    {
        using var client = CreateHttpClient();

        using var request = new HttpRequestMessage(HttpMethod.Post, LoginPath);
        AddTrustedForwardedHeaders(request);
        request.Headers.Add("Origin", PublicOrigin);
        request.Content = JsonContent.Create(new { password });

        return await client.SendAsync(request);
    }

    private static void AddTrustedForwardedHeaders(
        HttpRequestMessage request,
        string scheme = "https",
        string host = PublicHost)
    {
        request.Headers.Add("X-Forwarded-Proto", scheme);
        request.Headers.Add("X-Forwarded-For", "203.0.113.10");
        request.Headers.Add("X-Forwarded-Host", host);
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