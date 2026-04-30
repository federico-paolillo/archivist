using Archivist.Gateway.Api.Ping.Models;
using Archivist.Gateway.Application.Ping.Services;

using Microsoft.AspNetCore.Http.HttpResults;

namespace Archivist.Gateway.Api.Ping;

internal static class Handlers
{
    public static async Task<Ok<PongDto>> Ping(
        IPingService pingService
    )
    {
        var pong = pingService.Ping();

        var pongDto = new PongDto(pong.Time.ToUnixTimeSeconds());

        return TypedResults.Ok(pongDto);
    }
}