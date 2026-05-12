namespace Archivist.Gateway.Api.Articles.Models;

/// <summary>
/// Minimal public error response for article endpoints.
/// </summary>
internal sealed record ErrorResponse(string Error);