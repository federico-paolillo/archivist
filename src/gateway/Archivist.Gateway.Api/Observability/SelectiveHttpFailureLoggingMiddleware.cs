namespace Archivist.Gateway.Api.Observability;

using System.Diagnostics;

/// <summary>
/// Logs only security-relevant HTTP failures and operational server errors.
/// </summary>
public sealed partial class SelectiveHttpFailureLoggingMiddleware(
    RequestDelegate next,
    ILogger<SelectiveHttpFailureLoggingMiddleware> logger)
{
    public async Task InvokeAsync(HttpContext context)
    {
        ArgumentNullException.ThrowIfNull(context);

        try
        {
            await next(context).ConfigureAwait(false);
        }
        catch (Exception ex)
        {
            MarkCurrentActivityError("unhandled request exception");
            LogHttpRequestException(
                logger,
                ex,
                context.Request.Method,
                context.Request.Path.Value ?? string.Empty,
                context.GetEndpoint()?.DisplayName,
                context.Connection.RemoteIpAddress?.ToString(),
                Activity.Current?.TraceId.ToString(),
                Activity.Current?.SpanId.ToString());
            throw;
        }

        var statusCode = context.Response.StatusCode;
        if (statusCode >= StatusCodes.Status500InternalServerError)
        {
            MarkCurrentActivityError($"HTTP {statusCode}");
            LogHttpServerError(
                logger,
                context.Request.Method,
                context.Request.Path.Value ?? string.Empty,
                statusCode,
                context.GetEndpoint()?.DisplayName,
                context.Connection.RemoteIpAddress?.ToString(),
                Activity.Current?.TraceId.ToString(),
                Activity.Current?.SpanId.ToString());

            return;
        }

        if (statusCode is StatusCodes.Status401Unauthorized or StatusCodes.Status403Forbidden &&
            !IsRoutineAuthSessionProbe(context))
        {
            LogHttpSecurityFailure(
                logger,
                context.Request.Method,
                context.Request.Path.Value ?? string.Empty,
                statusCode,
                context.GetEndpoint()?.DisplayName,
                context.Connection.RemoteIpAddress?.ToString(),
                Activity.Current?.TraceId.ToString(),
                Activity.Current?.SpanId.ToString());
        }
    }

    private static bool IsRoutineAuthSessionProbe(HttpContext context) =>
        HttpMethods.IsGet(context.Request.Method) &&
        context.Request.Path.Equals("/auth/session", StringComparison.OrdinalIgnoreCase) &&
        context.Response.StatusCode == StatusCodes.Status401Unauthorized;

    private static void MarkCurrentActivityError(string description)
    {
        if (Activity.Current is { } activity)
        {
            activity.SetStatus(ActivityStatusCode.Error, description);
        }
    }

    [LoggerMessage(Level = LogLevel.Error, Message = "HTTP request failed with status {StatusCode}: {Method} {Path} endpoint {Endpoint} remote {RemoteIpAddress} trace {TraceId} span {SpanId}")]
    private static partial void LogHttpServerError(
        ILogger logger,
        string method,
        string path,
        int statusCode,
        string? endpoint,
        string? remoteIpAddress,
        string? traceId,
        string? spanId);

    [LoggerMessage(Level = LogLevel.Error, Message = "HTTP request threw an unhandled exception: {Method} {Path} endpoint {Endpoint} remote {RemoteIpAddress} trace {TraceId} span {SpanId}")]
    private static partial void LogHttpRequestException(
        ILogger logger,
        Exception exception,
        string method,
        string path,
        string? endpoint,
        string? remoteIpAddress,
        string? traceId,
        string? spanId);

    [LoggerMessage(Level = LogLevel.Warning, Message = "HTTP security-relevant response {StatusCode}: {Method} {Path} endpoint {Endpoint} remote {RemoteIpAddress} trace {TraceId} span {SpanId}")]
    private static partial void LogHttpSecurityFailure(
        ILogger logger,
        string method,
        string path,
        int statusCode,
        string? endpoint,
        string? remoteIpAddress,
        string? traceId,
        string? spanId);
}