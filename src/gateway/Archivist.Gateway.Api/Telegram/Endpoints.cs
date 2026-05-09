namespace Archivist.Gateway.Api.Telegram;

/// <summary>
/// Maps Telegram webhook routes to handler methods.
/// </summary>
internal static class Endpoints
{
    /// <summary>
    /// Registers the Telegram webhook route group.
    /// </summary>
    public static IEndpointRouteBuilder MapTelegram(this IEndpointRouteBuilder app)
    {
        var group = app.MapGroup("/telegram");

        group.MapPost("/webhook", Handlers.PostWebhook);

        return app;
    }
}