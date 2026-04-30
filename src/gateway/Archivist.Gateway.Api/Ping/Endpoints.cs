namespace Archivist.Gateway.Api.Ping;

internal static class Endpoints
{
    public static RouteGroupBuilder MapPing(this WebApplication webApplication)
    {
        var routeGroup = webApplication.MapGroup("/ping");

        routeGroup.MapGet("/", Handlers.Ping)
            .WithName("Ping");

        return routeGroup;
    }
}