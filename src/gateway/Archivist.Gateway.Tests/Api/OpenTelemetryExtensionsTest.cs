using Archivist.Gateway.Api.Observability;

using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Options;

using OpenTelemetry.Logs;

namespace Archivist.Gateway.Tests.Api;

public sealed class OpenTelemetryExtensionsTest
{
    [Fact]
    public void AddArchivistOpenTelemetry_RegistersTracingAndLoggingWhenExporterEnvVarsAreNone()
    {
        using var environment = new EnvironmentVariables(
            new("OTEL_TRACES_EXPORTER", "none"),
            new("OTEL_LOGS_EXPORTER", "none"),
            new("OTEL_EXPORTER_OTLP_ENDPOINT", "http://127.0.0.1:4318"));

        var services = new ServiceCollection();
        services.AddLogging();

        services.AddArchivistOpenTelemetry();

        Assert.Contains(
            services,
            static descriptor =>
                descriptor.ServiceType == typeof(IHostedService)
                && descriptor.ImplementationType?.FullName
                    == "OpenTelemetry.Extensions.Hosting.Implementation.TelemetryHostedService");
        Assert.Contains(
            services,
            static descriptor => descriptor.ServiceType.FullName == "OpenTelemetry.Trace.TracerProvider");
        Assert.Contains(
            services,
            static descriptor => descriptor.ServiceType.FullName == "OpenTelemetry.Logs.LoggerProvider");

        using var provider = services.BuildServiceProvider();

        Assert.Contains(
            provider.GetServices<ILoggerProvider>(),
            static loggerProvider => loggerProvider.GetType().FullName == "OpenTelemetry.Logs.OpenTelemetryLoggerProvider");

        var tracerProviderServiceType = Assert.Single(
            services,
            static descriptor => descriptor.ServiceType.FullName == "OpenTelemetry.Trace.TracerProvider")
            .ServiceType;
        Assert.NotNull(provider.GetService(tracerProviderServiceType));
    }

    [Fact]
    public void AddArchivistOpenTelemetryFilters_DropsLogsBelowInformationForOpenTelemetryProvider()
    {
        var services = new ServiceCollection();

        services.AddLogging(logging => logging.AddArchivistOpenTelemetryFilters());

        using var provider = services.BuildServiceProvider();
        var options = provider.GetRequiredService<IOptions<LoggerFilterOptions>>().Value;

        Assert.Contains(
            options.Rules,
            static rule =>
                rule.ProviderName == typeof(OpenTelemetryLoggerProvider).FullName
                && rule.CategoryName == "*"
                && rule.LogLevel == LogLevel.Information);
    }

    private sealed class EnvironmentVariables : IDisposable
    {
        private readonly Dictionary<string, string?> previousValues;

        public EnvironmentVariables(params EnvironmentVariable[] variables)
        {
            previousValues = variables.ToDictionary(
                static variable => variable.Name,
                static variable => Environment.GetEnvironmentVariable(variable.Name));

            foreach (var variable in variables)
            {
                Environment.SetEnvironmentVariable(variable.Name, variable.Value);
            }
        }

        public void Dispose()
        {
            foreach (var variable in previousValues)
            {
                Environment.SetEnvironmentVariable(variable.Key, variable.Value);
            }
        }
    }

    private sealed record EnvironmentVariable(string Name, string? Value);
}