namespace Archivist.Gateway.Application.Auth;

/// <summary>
/// Default values for the app-cookie authentication scheme.
/// </summary>
public static class AppCookieDefaults
{
    /// <summary>The default authentication scheme name for app-cookie auth.</summary>
    public const string AuthenticationScheme = "app-cookie";

    /// <summary>The default cookie name used for browser session auth.</summary>
    public const string CookieName = "__Host-app-auth";

    /// <summary>The fixed absolute lifetime for issued browser sessions.</summary>
    public static readonly TimeSpan SessionLifetime = TimeSpan.FromHours(24);
}