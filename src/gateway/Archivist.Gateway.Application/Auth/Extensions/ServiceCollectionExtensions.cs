using Archivist.Gateway.Application.Auth.Options;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Auth.Services.Defaults;

using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace Archivist.Gateway.Application.Auth.Extensions;

/// <summary>
/// Service registration extensions for the authentication feature.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Registers authentication services: password validation, Argon2id hashing, and auth bootstrap.
    /// </summary>
    public static IServiceCollection AddAuth(
        this IServiceCollection services,
        IConfiguration configuration)
    {
        ArgumentNullException.ThrowIfNull(services);
        ArgumentNullException.ThrowIfNull(configuration);

        services.Configure<AuthOptions>(opts =>
        {
            opts.SqlitePath = configuration["SQLITE_PATH"];
            opts.BootstrapPassword = configuration["AUTH_BOOTSTRAP_PASSWORD"];
        });

        services.AddSingleton<IPasswordValidator, PasswordValidator>();
        services.AddSingleton<IPasswordHasher, Argon2idPasswordHasher>();
        services.AddSingleton<IAuthBootstrapService, AuthBootstrapService>();

        return services;
    }
}