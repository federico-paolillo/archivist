namespace Archivist.Gateway.Api.Auth;

/// <summary>
/// Endpoint filter that rejects unsafe HTTP methods (POST, PUT, DELETE, PATCH) from cross-origin requests.
/// Requests must supply a valid <c>Origin</c> or <c>Referer</c> header that matches the effective request origin.
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
                return TypedResults.Forbid();
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
            return IsOriginSameRequestOrigin(originHeader, request, allowPath: false);
        }

        // Fall back to Referer header.
        var refererHeader = request.Headers.Referer.FirstOrDefault();
        if (!string.IsNullOrEmpty(refererHeader))
        {
            return IsOriginSameRequestOrigin(refererHeader, request, allowPath: true);
        }

        // Neither Origin nor Referer present: reject unsafe methods.
        return false;
    }

    private static bool IsOriginSameRequestOrigin(string origin, HttpRequest request, bool allowPath)
    {
        if (!Uri.TryCreate(origin, UriKind.Absolute, out var originUri))
        {
            return false;
        }

        if (!allowPath &&
            (originUri.AbsolutePath != "/" || !string.IsNullOrEmpty(originUri.Query) || !string.IsNullOrEmpty(originUri.Fragment)))
        {
            return false;
        }

        if (!string.Equals(originUri.Scheme, request.Scheme, StringComparison.OrdinalIgnoreCase))
        {
            return false;
        }

        if (!string.Equals(originUri.Host, request.Host.Host, StringComparison.OrdinalIgnoreCase))
        {
            return false;
        }

        return EffectivePort(originUri) == EffectiveRequestPort(request);
    }

    private static int EffectivePort(Uri uri)
    {
        if (!uri.IsDefaultPort)
        {
            return uri.Port;
        }

        return uri.Scheme.Equals("https", StringComparison.OrdinalIgnoreCase) ? 443 : 80;
    }

    private static int EffectiveRequestPort(HttpRequest request)
    {
        if (request.Host.Port is { } port)
        {
            return port;
        }

        return request.Scheme.Equals("https", StringComparison.OrdinalIgnoreCase) ? 443 : 80;
    }
}
