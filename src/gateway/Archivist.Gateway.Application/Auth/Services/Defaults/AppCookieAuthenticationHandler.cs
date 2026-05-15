using System.Security.Claims;
using System.Text.Encodings.Web;

using Archivist.Gateway.Application.Auth;

using Microsoft.AspNetCore.Authentication;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;

namespace Archivist.Gateway.Application.Auth.Services.Defaults;

/// <summary>
/// Custom ASP.NET Core authentication handler for the <c>app-cookie</c> scheme.
/// Reads the <c>__Host-app-auth</c> cookie, validates the session against <see cref="ISessionStore"/>,
/// and produces a minimal <see cref="ClaimsPrincipal"/> with one <see cref="ClaimTypes.NameIdentifier"/> claim.
/// The handler does not issue, clear, rotate, or refresh cookies.
/// </summary>
public sealed class AppCookieAuthenticationHandler(
    IOptionsMonitor<AppCookieSettings> options,
    ILoggerFactory logger,
    UrlEncoder encoder,
    ISessionStore sessionStore,
    TimeProvider timeProvider
) : AuthenticationHandler<AppCookieSettings>(options, logger, encoder)
{
    protected override async Task<AuthenticateResult> HandleAuthenticateAsync()
    {
        var cookieName = Options.CookieName;

        if (!Request.Cookies.TryGetValue(cookieName, out var sessionId) || string.IsNullOrEmpty(sessionId))
        {
            return AuthenticateResult.NoResult();
        }

        var entry = await sessionStore.GetAsync(sessionId, Context.RequestAborted);

        if (entry is null)
        {
            return AuthenticateResult.Fail("Session not found or expired.");
        }

        // Defensive expiry check: prune sessions whose absolute lifetime has passed.
        // InMemorySessionStore prunes eagerly inside GetAsync, but non-in-memory stores
        // (e.g., a future Redis implementation) may return entries past their TTL.
        if (timeProvider.GetUtcNow() >= entry.AbsoluteExpiresAt)
        {
            await sessionStore.RemoveAsync(sessionId, Context.RequestAborted);
            return AuthenticateResult.Fail("Session expired.");
        }

        var claims = new[]
        {
            new Claim(ClaimTypes.NameIdentifier, entry.UserId),
        };

        var identity = new ClaimsIdentity(claims, Scheme.Name);
        var principal = new ClaimsPrincipal(identity);
        var ticket = new AuthenticationTicket(principal, Scheme.Name);

        return AuthenticateResult.Success(ticket);
    }
}