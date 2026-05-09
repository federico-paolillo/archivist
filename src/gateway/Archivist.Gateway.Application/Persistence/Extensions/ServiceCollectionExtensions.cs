using Archivist.Gateway.Application.Persistence.Defaults;

using Microsoft.EntityFrameworkCore;
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
    public static IServiceCollection AddArchivistPersistence(this IServiceCollection services, string sqlitePath)
    {
        ArgumentException.ThrowIfNullOrWhiteSpace(sqlitePath);

        services.AddSingleton(TimeProvider.System);
        services.AddSingleton<IUlidGenerator, RandomUlidGenerator>();
        services.AddDbContext<ArchivistDbContext>(options => options.UseSqlite($"Data Source={sqlitePath}"));
        services.AddScoped<ITelegramIngestionRepository, EfTelegramIngestionRepository>();

        return services;
    }
}