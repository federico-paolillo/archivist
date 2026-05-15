using Archivist.Gateway.Application.Configuration;

namespace Archivist.Gateway.Api;

/// <summary>
/// Configures Gateway application configuration sources.
/// </summary>
public static class ConfigurationBuilderExtensions
{
    /// <summary>
    /// Uses the Gateway configuration source contract: optional appsettings defaults, then scoped environment variables.
    /// </summary>
    public static IConfigurationBuilder AddGatewayConfigurationSources(this IConfigurationBuilder configuration)
    {
        ArgumentNullException.ThrowIfNull(configuration);

        return configuration.AddEnvironmentVariables(Settings.EnvironmentVariablePrefix);
    }
}