using Archivist.Gateway.Api.Auth;
using Archivist.Gateway.Api.Ping;
using Archivist.Gateway.Api.Telegram;
using Archivist.Gateway.Application.Auth.Extensions;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Extensions;
using Archivist.Gateway.Application.Ping;
using Archivist.Gateway.Application.Telegram.Extensions;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddPing();
builder.Services.AddAuth(builder.Configuration);

var sqlitePath = builder.Configuration["SQLITE_PATH"];
if (!string.IsNullOrWhiteSpace(sqlitePath))
{
    builder.Services.AddArchivistPersistence(sqlitePath);
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

app.UseAuthentication();
app.UseAuthorization();

app.MapPing();
app.MapAuth();

if (!string.IsNullOrWhiteSpace(sqlitePath))
{
    app.MapTelegram();
}

await app.RunAsync();