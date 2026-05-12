using Archivist.Gateway.Api.Auth;

namespace Archivist.Gateway.Api.Articles;

/// <summary>
/// Maps article administration routes.
/// </summary>
internal static class Endpoints
{
    /// <summary>
    /// Registers article route mappings.
    /// </summary>
    public static IEndpointRouteBuilder MapArticles(this IEndpointRouteBuilder app)
    {
        app.MapDelete("/articles/{id}", Handlers.DeleteArticle)
            .RequireAuthorization()
            .AddEndpointFilter<SameOriginFilter>();

        return app;
    }
}