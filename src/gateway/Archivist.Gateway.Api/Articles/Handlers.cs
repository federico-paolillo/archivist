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
    /// GET /articles returns a fixed-size page of user-owned article metadata.
    /// </summary>
    public static async Task<Results<Ok<ArticleListResponse>, BadRequest<ErrorResponse>, UnauthorizedHttpResult>> ListArticles(
        string? after,
        string? before,
        HttpContext context,
        IArticleQueryService queryService,
        CancellationToken ct)
    {
        if (after is not null && before is not null)
        {
            return TypedResults.BadRequest(new ErrorResponse("Specify only one cursor."));
        }

        if (!TryNormalizeOptionalUlid(after, out var normalizedAfter) ||
            !TryNormalizeOptionalUlid(before, out var normalizedBefore))
        {
            return TypedResults.BadRequest(new ErrorResponse("Invalid article cursor."));
        }

        var userId = context.User.FindFirstValue(ClaimTypes.NameIdentifier);
        if (string.IsNullOrWhiteSpace(userId))
        {
            return TypedResults.Unauthorized();
        }

        var page = await queryService
            .ListAsync(userId, normalizedAfter, normalizedBefore, ct)
            .ConfigureAwait(false);

        return TypedResults.Ok(new ArticleListResponse(
            page.Items.Select(ToResponse).ToList(),
            page.NextCursor,
            page.PreviousCursor));
    }

    /// <summary>
    /// GET /articles/{id} returns user-owned article detail and Markdown artifacts.
    /// </summary>
    public static async Task<Results<Ok<ArticleDetailResponse>, BadRequest<ErrorResponse>, NotFound<ErrorResponse>, InternalServerError<ErrorResponse>, UnauthorizedHttpResult>> GetArticle(
        string id,
        HttpContext context,
        IArticleQueryService queryService,
        CancellationToken ct)
    {
        if (!TryNormalizeUlid(id, out var articleId))
        {
            return TypedResults.BadRequest(new ErrorResponse("Malformed article id."));
        }

        var userId = context.User.FindFirstValue(ClaimTypes.NameIdentifier);
        if (string.IsNullOrWhiteSpace(userId))
        {
            return TypedResults.Unauthorized();
        }

        var result = await queryService.GetDetailAsync(articleId, userId, ct).ConfigureAwait(false);

        return result.Kind switch
        {
            ArticleDetailResultKind.Found => TypedResults.Ok(ToResponse(result.Article!)),
            ArticleDetailResultKind.NotFound => TypedResults.NotFound(new ErrorResponse("Article not found.")),
            ArticleDetailResultKind.RequiredArtifactUnavailable => TypedResults.InternalServerError(new ErrorResponse("Required article artifacts are unavailable.")),
            _ => TypedResults.InternalServerError(new ErrorResponse("Article detail failed.")),
        };
    }

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

    /// <summary>
    /// DELETE /articles/{id}/force hard-deletes a user-owned article when running jobs are stale.
    /// </summary>
    public static async Task<Results<NoContent, BadRequest<ErrorResponse>, NotFound<ErrorResponse>, Conflict<ErrorResponse>, InternalServerError<ErrorResponse>>> ForceDeleteArticle(
        string id,
        HttpContext context,
        IArticleDeleteService deleteService,
        CancellationToken ct)
    {
        if (!TryNormalizeUlid(id, out var articleId))
        {
            return TypedResults.BadRequest(new ErrorResponse("Malformed article id."));
        }

        var userId = context.User.FindFirstValue(ClaimTypes.NameIdentifier);
        if (string.IsNullOrWhiteSpace(userId))
        {
            return TypedResults.NotFound(new ErrorResponse("Article not found."));
        }

        var result = await deleteService.ForceDeleteAsync(articleId, userId, ct).ConfigureAwait(false);

        return result switch
        {
            ArticleDeleteResult.Deleted => TypedResults.NoContent(),
            ArticleDeleteResult.NotFound => TypedResults.NotFound(new ErrorResponse("Article not found.")),
            ArticleDeleteResult.RunningJobConflict => TypedResults.Conflict(new ErrorResponse("Article has an active running job.")),
            ArticleDeleteResult.ArtifactCleanupFailed => TypedResults.InternalServerError(new ErrorResponse("Artifact cleanup failed.")),
            _ => TypedResults.InternalServerError(new ErrorResponse("Article force delete failed.")),
        };
    }

    private static ArticleListItemResponse ToResponse(ArticleListItem item) =>
        new(
            item.Id,
            item.Title,
            item.OriginalUrl,
            item.CanonicalUrl,
            item.Status,
            item.ErrorMessage,
            item.CreatedAt);

    private static ArticleDetailResponse ToResponse(ArticleDetail detail) =>
        new(
            detail.Id,
            detail.Title,
            detail.OriginalUrl,
            detail.CanonicalUrl,
            detail.Status,
            detail.ErrorMessage,
            detail.CreatedAt,
            detail.SummaryMarkdown,
            detail.ContentMarkdown,
            detail.CanForceDelete);

    private static bool TryNormalizeOptionalUlid(string? id, out string? normalized)
    {
        if (id is null)
        {
            normalized = null;
            return true;
        }

        var result = TryNormalizeUlid(id, out var normalizedId);
        normalized = normalizedId;
        return result;
    }

    private static bool TryNormalizeUlid(string id, out string normalized)
    {
        if (Ulid.TryParse(id, out var ulid))
        {
            normalized = ulid.ToString();
            return true;
        }

        normalized = string.Empty;
        return false;
    }
}