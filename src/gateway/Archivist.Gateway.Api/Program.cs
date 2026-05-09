using Archivist.Gateway.Api.Ping;
using Archivist.Gateway.Application.Auth.Extensions;
using Archivist.Gateway.Application.Auth.Services;
using Archivist.Gateway.Application.Ping;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddPing();
builder.Services.AddAuth(builder.Configuration);

var app = builder.Build();

// Run auth bootstrap before accepting requests.
// If bootstrap fails the application will not start.
var authBootstrap = app.Services.GetRequiredService<IAuthBootstrapService>();
await authBootstrap.InitializeAsync();

app.MapPing();

await app.RunAsync();