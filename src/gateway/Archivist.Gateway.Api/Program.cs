using Archivist.Gateway.Api.Ping;
using Archivist.Gateway.Application.Persistence;
using Archivist.Gateway.Application.Persistence.Extensions;
using Archivist.Gateway.Application.Ping;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddPing();

var sqlitePath = builder.Configuration["SQLITE_PATH"];
if (!string.IsNullOrWhiteSpace(sqlitePath))
{
    builder.Services.AddArchivistPersistence(sqlitePath);
}

var app = builder.Build();

if (!string.IsNullOrWhiteSpace(sqlitePath))
{
    using var scope = app.Services.CreateScope();
    var db = scope.ServiceProvider.GetRequiredService<ArchivistDbContext>();
    await db.Database.EnsureCreatedAsync();
}

app.MapPing();

await app.RunAsync();