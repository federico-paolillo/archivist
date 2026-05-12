using System.Security.Claims;

using Archivist.Gateway.Api.Articles.Models;
using Archivist.Gateway.Application.Articles;

using Microsoft.AspNetCore.Http.HttpResults;

using NUlid;

namespace Archivist.Gateway.Api.Articles;

/// <summary>
/// Static handler methods for article endpoints.
/// </summary>
internal static class Handlers
{
    /// <summary>
    /// DELETE /articles/{id} hard-deletes a user-owned article.
    /// </summary>
    public static async Task<Results<NoContent, BadRequest<ErrorResponse>, NotFound<ErrorResponse>, Conflict<ErrorResponse>, InternalServerError<ErrorResponse>>> DeleteArticle(
        string id,
        HttpContext context,
        IArticleDeleteService deleteService,
        CancellationToken ct)
    {
        if (!Ulid.TryParse(id, out _))
        {
            return TypedResults.BadRequest(new ErrorResponse("Malformed article id."));
        }

        var userId = context.User.FindFirstValue(ClaimTypes.NameIdentifier);
        if (string.IsNullOrWhiteSpace(userId))
        {
            return TypedResults.NotFound(new ErrorResponse("Article not found."));
        }

        var result = await deleteService.DeleteAsync(id, userId, ct).ConfigureAwait(false);

        return result switch
        {
            ArticleDeleteResult.Deleted => TypedResults.NoContent(),
            ArticleDeleteResult.NotFound => TypedResults.NotFound(new ErrorResponse("Article not found.")),
            ArticleDeleteResult.RunningJobConflict => TypedResults.Conflict(new ErrorResponse("Article has a running job.")),
            ArticleDeleteResult.ArtifactCleanupFailed => TypedResults.InternalServerError(new ErrorResponse("Artifact cleanup failed.")),
            _ => TypedResults.InternalServerError(new ErrorResponse("Article delete failed.")),
        };
    }
}