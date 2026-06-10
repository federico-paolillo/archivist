namespace Archivist.Gateway.Tests.Api;

using System.Diagnostics;
using System.Net;

using Archivist.Gateway.Api.Observability;

using Microsoft.AspNetCore.Http;
using Microsoft.Extensions.Logging;

public sealed class SelectiveHttpFailureLoggingMiddlewareTest
{
    [Fact]
    public async Task InvokeAsync_ServerError_LogsErrorAndMarksCurrentActivityError()
    {
        var logger = new CapturingLogger<SelectiveHttpFailureLoggingMiddleware>();
        var middleware = new SelectiveHttpFailureLoggingMiddleware(
            context =>
            {
                context.Response.StatusCode = StatusCodes.Status500InternalServerError;
                return Task.CompletedTask;
            },
            logger);
        var context = CreateContext(HttpMethods.Get, "/articles", "?token=secret");

        using var activity = new Activity("test-request");
        activity.Start();

        await middleware.InvokeAsync(context);

        var entry = Assert.Single(logger.Entries);
        Assert.Equal(LogLevel.Error, entry.Level);
        Assert.Contains("GET /articles", entry.Message, StringComparison.Ordinal);
        Assert.Contains("500", entry.Message, StringComparison.Ordinal);
        Assert.DoesNotContain("token=secret", entry.Message, StringComparison.Ordinal);
        Assert.Equal(ActivityStatusCode.Error, activity.Status);
    }

    [Theory]
    [InlineData(StatusCodes.Status401Unauthorized)]
    [InlineData(StatusCodes.Status403Forbidden)]
    public async Task InvokeAsync_SecurityRelevantAuthFailure_LogsWarning(int statusCode)
    {
        var logger = new CapturingLogger<SelectiveHttpFailureLoggingMiddleware>();
        var middleware = new SelectiveHttpFailureLoggingMiddleware(
            context =>
            {
                context.Response.StatusCode = statusCode;
                return Task.CompletedTask;
            },
            logger);
        var context = CreateContext(HttpMethods.Post, "/login");

        await middleware.InvokeAsync(context);

        var entry = Assert.Single(logger.Entries);
        Assert.Equal(LogLevel.Warning, entry.Level);
        Assert.Contains("POST /login", entry.Message, StringComparison.Ordinal);
        Assert.Contains(statusCode.ToStringInvariant(), entry.Message, StringComparison.Ordinal);
    }

    [Fact]
    public async Task InvokeAsync_RoutineUnauthenticatedSessionProbe_DoesNotLog()
    {
        var logger = new CapturingLogger<SelectiveHttpFailureLoggingMiddleware>();
        var middleware = new SelectiveHttpFailureLoggingMiddleware(
            context =>
            {
                context.Response.StatusCode = StatusCodes.Status401Unauthorized;
                return Task.CompletedTask;
            },
            logger);
        var context = CreateContext(HttpMethods.Get, "/auth/session");

        await middleware.InvokeAsync(context);

        Assert.Empty(logger.Entries);
    }

    [Theory]
    [InlineData(StatusCodes.Status200OK)]
    [InlineData(StatusCodes.Status404NotFound)]
    public async Task InvokeAsync_NonSelectedStatus_DoesNotLog(int statusCode)
    {
        var logger = new CapturingLogger<SelectiveHttpFailureLoggingMiddleware>();
        var middleware = new SelectiveHttpFailureLoggingMiddleware(
            context =>
            {
                context.Response.StatusCode = statusCode;
                return Task.CompletedTask;
            },
            logger);
        var context = CreateContext(HttpMethods.Get, "/articles");

        await middleware.InvokeAsync(context);

        Assert.Empty(logger.Entries);
    }

    private static DefaultHttpContext CreateContext(string method, string path, string query = "")
    {
        var context = new DefaultHttpContext();
        context.Request.Method = method;
        context.Request.Path = path;
        context.Request.QueryString = new QueryString(query);
        context.Connection.RemoteIpAddress = IPAddress.Loopback;
        return context;
    }

    private sealed class CapturingLogger<T> : ILogger<T>
    {
        public List<LogEntry> Entries { get; } = [];

        public IDisposable? BeginScope<TState>(TState state)
            where TState : notnull =>
            null;

        public bool IsEnabled(LogLevel logLevel) => true;

        public void Log<TState>(
            LogLevel logLevel,
            EventId eventId,
            TState state,
            Exception? exception,
            Func<TState, Exception?, string> formatter)
        {
            Entries.Add(new LogEntry(logLevel, formatter(state, exception)));
        }
    }

    private sealed record LogEntry(LogLevel Level, string Message);
}

internal static class IntFormattingExtensions
{
    public static string ToStringInvariant(this int value) =>
        value.ToString(System.Globalization.CultureInfo.InvariantCulture);
}