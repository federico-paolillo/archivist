using Archivist.Gateway.Application.Configuration;
using Archivist.Gateway.Application.Persistence.Defaults;

using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;

namespace Archivist.Gateway.Application.Persistence.Extensions;

/// <summary>
/// Registers gateway persistence services.
/// </summary>
public static class ServiceCollectionExtensions
{
    /// <summary>
    /// Adds SQLite-backed Archivist persistence.
    /// </summary>
    public static IServiceCollection AddArchivistPersistence(this IServiceCollection services, IConfiguration configuration)
    {
        ArgumentNullException.ThrowIfNull(services);
        ArgumentNullException.ThrowIfNull(configuration);

        services.AddSingleton(TimeProvider.System);
        services.AddSingleton<IUlidGenerator, RandomUlidGenerator>();
        services.AddDbContext<ArchivistDbContext>((serviceProvider, options) =>
        {
            var resolvedConfiguration = serviceProvider.GetRequiredService<IConfiguration>();
            var sqlitePath = resolvedConfiguration.GetValue<string>(Settings.SqlitePathKey);

            ArgumentException.ThrowIfNullOrWhiteSpace(sqlitePath);

            options.UseSqlite($"Data Source={sqlitePath}");
        });
        services.AddScoped<ITelegramIngestionRepository, EfTelegramIngestionRepository>();

        return services;
    }
}