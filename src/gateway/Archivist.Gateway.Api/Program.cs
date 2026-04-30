using Archivist.Gateway.Api.Ping;
using Archivist.Gateway.Application.Ping;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddPing();

var app = builder.Build();

app.MapPing();

app.Run();