using System;

using Archivist.Gateway.Application.Configuration;

using Microsoft.AspNetCore.HttpOverrides;
using Microsoft.Extensions.Options;

namespace Archivist.Gateway.Api.Auth;

public static class ServiceCollectionExtensions
{
    public static IServiceCollection AddForwardedHeaders(
      this IServiceCollection serviceCollection,
      IConfiguration configuration,
      IHostEnvironment hostEnvironment
    )
    {
        ArgumentNullException.ThrowIfNull(configuration);
        ArgumentNullException.ThrowIfNull(hostEnvironment);

        serviceCollection
            .AddOptions<ForwardedHeadersOptions>()
            .Configure<IConfiguration, IHostEnvironment>((options, resolvedConfiguration, resolvedEnvironment) =>
            {
                var publicHosts = GetGatewayPublicHosts(resolvedConfiguration, resolvedEnvironment);

                options.ForwardedHeaders = ForwardedHeaders.XForwardedFor | ForwardedHeaders.XForwardedProto | ForwardedHeaders.XForwardedHost;
                options.ForwardLimit = 1;
                options.KnownIPNetworks.Clear();
                options.KnownProxies.Clear();

                foreach (var host in publicHosts)
                {
                    options.AllowedHosts.Add(host);
                }
            });

        return serviceCollection;
    }

    private static string[] GetGatewayPublicHosts(
      IConfiguration configuration,
      IHostEnvironment environment
    )
    {
        var configuredHosts = configuration.GetValue<string>(Settings.GatewayPublicHostsKey);

        var publicHosts = configuredHosts?
                .Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries)
                .Distinct(StringComparer.OrdinalIgnoreCase)
                .ToArray() ??
            [];

        if (!environment.IsDevelopment() && publicHosts.Length == 0)
        {
            throw new InvalidOperationException($"{Settings.GatewayPublicHostsKey} is required outside Development.");
        }

        return publicHosts;
    }
}