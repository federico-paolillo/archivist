using System.Security.Cryptography;

using Archivist.Gateway.Api.Auth.Models;
using Archivist.Gateway.Application.Auth;
using Archivist.Gateway.Application.Auth.Options;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Persistence;

using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Options;

namespace Archivist.Gateway.Api.Auth;

/// <summary>
/// Static handler methods for auth endpoints: login, logout, and session check.
/// </summary>
internal static class Handlers
{
    private const int SessionIdByteLength = 32;

    /// <summary>
    /// POST /login — verify password, issue session cookie, return 204. Returns 401 on failure.
    /// </summary>
    public static async Task<IResult> PostLogin(
        HttpContext context,
        LoginRequest? body,
        IPasswordStore passwordStore,
        IPasswordHasher passwordHasher,
        IPasswordValidator passwordValidator,
        ISessionStore sessionStore,
        ILoginThrottle loginThrottle,
        IOptions<AppCookieOptions> cookieOptionsAccessor,
        TimeProvider timeProvider,
        CancellationToken ct)
    {
        var cookieOptions = cookieOptionsAccessor.Value;

        if (!context.Request.Scheme.Equals("https", StringComparison.OrdinalIgnoreCase))
        {
            return TypedResults.Forbid();
        }

        // Reject if body or password field is missing.
        var password = body?.Password;
        if (string.IsNullOrEmpty(password))
        {
            return TypedResults.Unauthorized();
        }

        // Reject malformed or oversized password before Argon2id work.
        if (!passwordValidator.IsValid(password))
        {
            return TypedResults.Unauthorized();
        }

        var sourceIp = context.Connection.RemoteIpAddress?.ToString() ?? "unknown";

        // Apply throttle check before Argon2id verification.
        if (loginThrottle.IsThrottled(sourceIp))
        {
            return TypedResults.Unauthorized();
        }

        // Read stored hash. Return generic 401 if none exists.
        var storedHash = await passwordStore.GetPasswordHashAsync(ct);
        if (storedHash is null)
        {
            loginThrottle.RecordFailure(sourceIp);
            return TypedResults.Unauthorized();
        }

        // Verify password with Argon2id constant-time comparison.
        if (!passwordHasher.Verify(password, storedHash))
        {
            loginThrottle.RecordFailure(sourceIp);
            return TypedResults.Unauthorized();
        }

        // Successful login — rotate session if an existing valid session is present.
        var existingSessionId = context.Request.Cookies[cookieOptions.CookieName];
        if (!string.IsNullOrEmpty(existingSessionId))
        {
            await sessionStore.RemoveAsync(existingSessionId, ct);
        }

        // Generate fresh 32-byte CSPRNG session id, base64url-encoded without padding.
        var sessionIdBytes = RandomNumberGenerator.GetBytes(SessionIdByteLength);
        var newSessionId = Convert.ToBase64String(sessionIdBytes)
            .Replace('+', '-')
            .Replace('/', '_')
            .TrimEnd('=');

        var personalUserId = PersistenceConstants.PersonalUserId;

        var now = timeProvider.GetUtcNow();
        var entry = new SessionEntry(
            UserId: personalUserId,
            CreatedAt: now,
            AbsoluteExpiresAt: now.Add(cookieOptions.SessionLifetime));

        await sessionStore.SetAsync(newSessionId, entry, ct);

        loginThrottle.RecordSuccess(sourceIp);

        // Set __Host-app-auth cookie with required attributes.
        // __Host- prefix requires: Secure, Path=/, no Domain.
        // REQ-011: no Expires or Max-Age on login.
        context.Response.Cookies.Append(
            cookieOptions.CookieName,
            newSessionId,
            new CookieOptions
            {
                HttpOnly = true,
                Secure = true,
                SameSite = SameSiteMode.Strict,
                Path = "/",
                // Domain must not be set for __Host- prefix compliance.
                IsEssential = true,
            });

        return TypedResults.NoContent();
    }

    /// <summary>
    /// POST /logout — remove session entry if present, always clear cookie, return 204.
    /// </summary>
    public static async Task<NoContent> PostLogout(
        HttpContext context,
        ISessionStore sessionStore,
        IOptions<AppCookieOptions> cookieOptionsAccessor,
        CancellationToken ct)
    {
        var cookieOptions = cookieOptionsAccessor.Value;
        var sessionId = context.Request.Cookies[cookieOptions.CookieName];

        if (!string.IsNullOrEmpty(sessionId))
        {
            await sessionStore.RemoveAsync(sessionId, ct);
        }

        // Always clear the cookie with Max-Age=0.
        context.Response.Cookies.Append(
            cookieOptions.CookieName,
            string.Empty,
            new CookieOptions
            {
                HttpOnly = true,
                Secure = true,
                SameSite = SameSiteMode.Strict,
                Path = "/",
                MaxAge = TimeSpan.Zero,
            });

        return TypedResults.NoContent();
    }

    /// <summary>
    /// GET /auth/session — return 204 if authenticated, 401 otherwise.
    /// </summary>
    public static IResult GetSession(HttpContext context)
    {
        return context.User.Identity?.IsAuthenticated == true
            ? TypedResults.NoContent()
            : TypedResults.Unauthorized();
    }
}