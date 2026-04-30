using Archivist.Gateway.Application.Ping.Services;
using Archivist.Gateway.Application.Ping.Services.Defaults;

using Microsoft.Extensions.DependencyInjection;

namespace Archivist.Gateway.Application.Ping;

public static class ServiceCOllectionExtensions
{
    public static IServiceCollection AddPing(this IServiceCollection serviceCollection)
    {
        ArgumentNullException.ThrowIfNull(serviceCollection);

        serviceCollection.AddSingleton(TimeProvider.System);
        serviceCollection.AddSingleton<IPingService, PingService>();

        return serviceCollection;
    }
}