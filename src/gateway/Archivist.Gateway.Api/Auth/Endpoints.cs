namespace Archivist.Gateway.Api.Auth;

/// <summary>
/// Maps authentication routes: POST /login, POST /logout, GET /auth/session.
/// </summary>
internal static class Endpoints
{
    /// <summary>
    /// Registers the auth route group.
    /// </summary>
    public static IEndpointRouteBuilder MapAuth(this IEndpointRouteBuilder app)
    {
        app.MapPost("/login", Handlers.PostLogin)
            .AddEndpointFilter<SameOriginFilter>();

        app.MapPost("/logout", Handlers.PostLogout)
            .AddEndpointFilter<SameOriginFilter>();

        app.MapGet("/auth/session", Handlers.GetSession)
            .RequireAuthorization();

        return app;
    }
}