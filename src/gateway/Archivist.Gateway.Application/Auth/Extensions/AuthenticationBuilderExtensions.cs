using Archivist.Gateway.Application.Auth.Options;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.AspNetCore.Authentication;

namespace Archivist.Gateway.Application.Auth.Extensions;

/// <summary>
/// Extension methods for registering the app-cookie authentication scheme.
/// </summary>
public static class AuthenticationBuilderExtensions
{
    /// <summary>
    /// Registers the <c>app-cookie</c> authentication scheme using the custom <see cref="AppCookieAuthenticationHandler"/>.
    /// Default scheme name: <c>"app-cookie"</c>.
    /// Default cookie name: <c>"__Host-app-auth"</c>.
    /// Default session lifetime: 24 hours.
    /// </summary>
    public static AuthenticationBuilder AddAppCookie(
        this AuthenticationBuilder builder,
        Action<AppCookieOptions>? configure = null)
    {
        ArgumentNullException.ThrowIfNull(builder);

        return builder.AddScheme<AppCookieOptions, AppCookieAuthenticationHandler>(
            AppCookieDefaults.AuthenticationScheme,
            configure);
    }
}