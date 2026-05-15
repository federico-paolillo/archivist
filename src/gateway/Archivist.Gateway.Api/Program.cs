using Archivist.Gateway.Api.Articles;
using Archivist.Gateway.Api.Auth;
using Archivist.Gateway.Api.Ping;
using Archivist.Gateway.Api.Telegram;
using Archivist.Gateway.Application.Articles.Extensions;
using Archivist.Gateway.Application.Auth.Extensions;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Extensions;
using Archivist.Gateway.Application.Ping;
using Archivist.Gateway.Application.Telegram.Extensions;

using Microsoft.AspNetCore.HttpOverrides;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddPing();
builder.Services.AddAuth(builder.Configuration);

var publicHosts = GetGatewayPublicHosts(builder.Configuration, builder.Environment);
builder.Services.Configure<ForwardedHeadersOptions>(options =>
{
    options.ForwardedHeaders = ForwardedHeaders.XForwardedFor |
        ForwardedHeaders.XForwardedProto |
        ForwardedHeaders.XForwardedHost;
    options.ForwardLimit = 1;
    options.KnownIPNetworks.Clear();
    options.KnownProxies.Clear();

    foreach (var host in publicHosts)
    {
        options.AllowedHosts.Add(host);
    }
});

var sqlitePath = builder.Configuration["SQLITE_PATH"];
if (!string.IsNullOrWhiteSpace(sqlitePath))
{
    builder.Services.AddArchivistPersistence(sqlitePath);
    builder.Services.AddArticles(builder.Configuration);
    builder.Services.AddTelegram(builder.Configuration);
}

var app = builder.Build();

if (!string.IsNullOrWhiteSpace(sqlitePath))
{
    using var scope = app.Services.CreateScope();
    var db = scope.ServiceProvider.GetRequiredService<ArchivistDbContext>();
    await db.Database.EnsureCreatedAsync();
}

// Run auth bootstrap before accepting requests.
// If bootstrap fails the application will not start.
var authBootstrap = app.Services.GetRequiredService<IAuthBootstrapService>();
await authBootstrap.InitializeAsync();

app.UseForwardedHeaders();
app.UseAuthentication();
app.UseAuthorization();

app.MapPing();
app.MapAuth();

if (!string.IsNullOrWhiteSpace(sqlitePath))
{
    app.MapArticles();
    app.MapTelegram();
}

await app.RunAsync();

static IReadOnlyList<string> GetGatewayPublicHosts(IConfiguration configuration, IHostEnvironment environment)
{
    var configuredHosts = configuration["GATEWAY_PUBLIC_HOSTS"];
    var publicHosts = configuredHosts?
            .Split(',', StringSplitOptions.RemoveEmptyEntries | StringSplitOptions.TrimEntries)
            .Distinct(StringComparer.OrdinalIgnoreCase)
            .ToArray() ??
        [];

    if (!environment.IsDevelopment() && publicHosts.Length == 0)
    {
        throw new InvalidOperationException("GATEWAY_PUBLIC_HOSTS is required outside Development.");
    }

    return publicHosts;
}
