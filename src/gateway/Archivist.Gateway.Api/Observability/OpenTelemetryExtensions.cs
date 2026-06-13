using Archivist.Gateway.Application.Observability;

using OpenTelemetry.Exporter;
using OpenTelemetry.Logs;
using OpenTelemetry.Resources;
using OpenTelemetry.Trace;

namespace Archivist.Gateway.Api.Observability;

/// <summary>
/// Registers Gateway OpenTelemetry tracing and logging.
/// </summary>
public static class OpenTelemetryExtensions
{
    public static ILoggingBuilder AddArchivistOpenTelemetryFilters(this ILoggingBuilder logging)
    {
        ArgumentNullException.ThrowIfNull(logging);

        logging.AddFilter<OpenTelemetryLoggerProvider>("*", LogLevel.Information);

        return logging;
    }

    public static IServiceCollection AddArchivistOpenTelemetry(this IServiceCollection services)
    {
        ArgumentNullException.ThrowIfNull(services);

        services.AddOpenTelemetry()
            .ConfigureResource(resource => resource.AddService(ArchivistTelemetry.ServiceName))
            .WithTracing(tracing =>
            {
                tracing
                    .AddSource(ArchivistTelemetry.ActivitySourceName)
                    .AddAspNetCoreInstrumentation(options =>
                    {
                        options.RecordException = true;
                        options.Filter = context => !context.Request.Path.StartsWithSegments("/ping");
                    })
                    .AddHttpClientInstrumentation(options =>
                    {
                        options.FilterHttpRequestMessage = request =>
                            request.RequestUri?.Host is not "api.telegram.org";
                    })
                    .AddOtlpExporter(options => options.Protocol = OtlpExportProtocol.HttpProtobuf);
            })
            .WithLogging(logging =>
            {
                logging.AddOtlpExporter(options => options.Protocol = OtlpExportProtocol.HttpProtobuf);
            });

        return services;
    }
}