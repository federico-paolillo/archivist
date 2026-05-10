using Microsoft.AspNetCore.Authentication;

namespace Archivist.Gateway.Application.Auth.Options;

/// <summary>
/// Configuration options for the app-cookie authentication scheme.
/// </summary>
public sealed class AppCookieOptions : AuthenticationSchemeOptions
{
    /// <summary>The name of the authentication cookie. Defaults to "__Host-app-auth".</summary>
    public string CookieName { get; set; } = "__Host-app-auth";

    /// <summary>The absolute lifetime of an issued session. Defaults to 24 hours.</summary>
    public TimeSpan SessionLifetime { get; set; } = TimeSpan.FromHours(24);
}