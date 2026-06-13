using Archivist.Gateway.Api;
using Archivist.Gateway.Api.Articles;
using Archivist.Gateway.Api.Auth;
using Archivist.Gateway.Api.Observability;
using Archivist.Gateway.Api.Ping;
using Archivist.Gateway.Api.Telegram;
using Archivist.Gateway.Application.Articles.Extensions;
using Archivist.Gateway.Application.Auth.Extensions;
using Archivist.Gateway.Application.Persistence.Extensions;
using Archivist.Gateway.Application.Ping;
using Archivist.Gateway.Application.Telegram.Extensions;

using Microsoft.Extensions.Logging.Console;

var builder = WebApplication.CreateBuilder();

builder.Configuration.AddGatewayConfigurationSources();

builder.Logging.ClearProviders();
builder.Logging.AddSimpleConsole(options =>
    {
        options.SingleLine = true;
        options.TimestampFormat = "HH:mm:ss ";
        options.UseUtcTimestamp = true;
        options.ColorBehavior = LoggerColorBehavior.Disabled;
        options.IncludeScopes = true;
    }
);

builder.Logging.AddFilter("Microsoft.EntityFramework", LogLevel.Warning);
builder.Logging.AddArchivistOpenTelemetryFilters();

builder.Services.AddAuth(builder.Configuration);
builder.Services.AddForwardedHeaders(builder.Configuration, builder.Environment);
builder.Services.AddArchivistPersistence(builder.Configuration);
builder.Services.AddArchivistOpenTelemetry();

builder.Services.AddPing();
builder.Services.AddArticles(builder.Configuration);
builder.Services.AddTelegram(builder.Configuration);

var app = builder.Build();

app.UseForwardedHeaders();
app.UseMiddleware<SelectiveHttpFailureLoggingMiddleware>();
app.UseAuthentication();
app.UseAuthorization();

app.MapPing();
app.MapAuth();
app.MapArticles();
app.MapTelegram();

await app.PrepareAsync();

await app.RunAsync();