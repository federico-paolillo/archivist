namespace Archivist.Gateway.Api.Auth;

/// <summary>
/// Endpoint filter that rejects unsafe HTTP methods (POST, PUT, DELETE, PATCH) from cross-origin requests.
/// Requests must supply a valid <c>Origin</c> or <c>Referer</c> header that matches the request host.
/// </summary>
[System.Diagnostics.CodeAnalysis.SuppressMessage("Performance", "CA1812:Avoid uninstantiated internal classes", Justification = "Instantiated via AddEndpointFilter<T>()")]
internal sealed class SameOriginFilter : IEndpointFilter
{
    private static readonly HashSet<string> UnsafeMethods = new(StringComparer.OrdinalIgnoreCase)
    {
        HttpMethods.Post,
        HttpMethods.Put,
        HttpMethods.Delete,
        HttpMethods.Patch,
    };

    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        var httpContext = context.HttpContext;
        var request = httpContext.Request;

        if (UnsafeMethods.Contains(request.Method))
        {
            if (!IsSameOrigin(request))
            {
                return TypedResults.StatusCode(StatusCodes.Status403Forbidden);
            }
        }

        return await next(context);
    }

    private static bool IsSameOrigin(HttpRequest request)
    {
        // Check Origin header first.
        var originHeader = request.Headers.Origin.FirstOrDefault();
        if (!string.IsNullOrEmpty(originHeader))
        {
            return IsOriginSameHost(originHeader, request);
        }

        // Fall back to Referer header.
        var refererHeader = request.Headers.Referer.FirstOrDefault();
        if (!string.IsNullOrEmpty(refererHeader))
        {
            return IsRefererSameHost(refererHeader, request);
        }

        // Neither Origin nor Referer present: reject unsafe methods.
        return false;
    }

    private static bool IsOriginSameHost(string origin, HttpRequest request)
    {
        if (!Uri.TryCreate(origin, UriKind.Absolute, out var originUri))
        {
            return false;
        }

        var requestHost = request.Host.Host;
        var requestPort = request.Host.Port;

        if (!string.Equals(originUri.Host, requestHost, StringComparison.OrdinalIgnoreCase))
        {
            return false;
        }

        // Check port: origin port must match the request's effective port.
        var originPort = originUri.IsDefaultPort ? -1 : originUri.Port;
        var effectiveRequestPort = requestPort ?? (request.IsHttps ? 443 : 80);

        // Normalize default ports.
        if (originUri.Scheme.Equals("https", StringComparison.OrdinalIgnoreCase) && originPort == -1)
        {
            originPort = 443;
        }
        else if (originUri.Scheme.Equals("http", StringComparison.OrdinalIgnoreCase) && originPort == -1)
        {
            originPort = 80;
        }

        return originPort == effectiveRequestPort;
    }

    private static bool IsRefererSameHost(string referer, HttpRequest request)
    {
        if (!Uri.TryCreate(referer, UriKind.Absolute, out var refererUri))
        {
            return false;
        }

        return IsOriginSameHost(
            $"{refererUri.Scheme}://{refererUri.Authority}",
            request);
    }
}