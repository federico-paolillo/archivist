using Archivist.Gateway.Application.Auth;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;
using Archivist.Gateway.Application.Configuration;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace Archivist.Gateway.Application.Auth.Extensions;

/// <summary>
/// Service registration extensions for the authentication feature.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Registers authentication services: password validation, Argon2id hashing, auth bootstrap,
    /// session store, login throttle, password store, and the app-cookie authentication handler.
    /// </summary>
    public static IServiceCollection AddAuth(
        this IServiceCollection services,
        IConfiguration configuration)
    {
        ArgumentNullException.ThrowIfNull(services);
        ArgumentNullException.ThrowIfNull(configuration);

        services.AddSingleton(serviceProvider =>
        {
            var resolvedConfiguration = serviceProvider.GetRequiredService<IConfiguration>();

            return new AuthSettings
            {
                SqlitePath = resolvedConfiguration.GetValue<string>(Settings.SqlitePathKey),
                BootstrapPassword = resolvedConfiguration.GetValue<string>(Settings.AuthBootstrapPasswordKey),
            };
        });

        services.AddSingleton<IPasswordValidator, PasswordValidator>();
        services.AddSingleton<IPasswordHasher, Argon2idPasswordHasher>();
        services.AddSingleton<IAuthBootstrapService, AuthBootstrapService>();
        services.AddSingleton<IPasswordStore, SqlitePasswordStore>();
        services.AddSingleton<ISessionStore, InMemorySessionStore>();
        services.AddSingleton<ILoginThrottle, InMemoryLoginThrottle>();

        // Register TimeProvider for InMemorySessionStore and other consumers.
        services.AddSingleton(TimeProvider.System);

        services
            .AddAuthentication(AppCookieDefaults.AuthenticationScheme)
            .AddAppCookie();

        services.AddAuthorization();

        return services;
    }

}