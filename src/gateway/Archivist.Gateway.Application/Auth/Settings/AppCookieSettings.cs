using Microsoft.AspNetCore.Authentication;

namespace Archivist.Gateway.Application.Auth;

/// <summary>
/// Configuration settings for the app-cookie authentication scheme.
/// </summary>
public sealed class AppCookieSettings : AuthenticationSchemeOptions
{
    public const string Section = global::Archivist.Gateway.Application.Configuration.Settings.AppCookieSection;

    /// <summary>The name of the authentication cookie. Defaults to "__Host-app-auth".</summary>
    public string CookieName { get; set; } = "__Host-app-auth";

    /// <summary>The absolute lifetime of an issued session. Defaults to 24 hours.</summary>
    public TimeSpan SessionLifetime { get; set; } = TimeSpan.FromHours(24);
}