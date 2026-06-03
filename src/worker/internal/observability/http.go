package observability

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/imroc/req/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func ReqRoundTripWrapper() req.RoundTripWrapperFunc {
	return func(rt req.RoundTripper) req.RoundTripFunc {
		return func(request *req.Request) (*req.Response, error) {
			method, methodAttrs := httpMethodAttributes(request.Method)
			ctx, span := Tracer().Start(
				request.Context(),
				"HTTP "+method,
				trace.WithSpanKind(trace.SpanKindClient),
				trace.WithAttributes(append(methodAttrs, urlAttributes(request.URL)...)...),
			)
			request.SetContext(ctx)
			injectTraceContext(request)

			var err error
			defer func() {
				EndSpan(span, err)
			}()

			response, err := rt.RoundTrip(request)
			if response != nil && response.Response != nil {
				statusCode := response.GetStatusCode()
				span.SetAttributes(semconv.HTTPResponseStatusCode(statusCode))

				if statusCode >= http.StatusInternalServerError {
					span.SetStatus(codes.Error, response.GetStatus())
				}
			}

			return response, err //nolint:wrapcheck // Middleware must preserve req's transport error contract.
		}
	}
}

func injectTraceContext(request *req.Request) {
	if request.Headers == nil {
		request.Headers = http.Header{}
	}

	otel.GetTextMapPropagator().Inject(request.Context(), propagation.HeaderCarrier(request.Headers))
}

func httpMethodAttributes(method string) (string, []attribute.KeyValue) {
	upperMethod := strings.ToUpper(method)
	switch upperMethod {
	case http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace:
		attrs := []attribute.KeyValue{semconv.HTTPRequestMethodKey.String(upperMethod)}
		if method != upperMethod {
			attrs = append(attrs, semconv.HTTPRequestMethodOriginal(method))
		}

		return upperMethod, attrs
	default:
		attrs := []attribute.KeyValue{semconv.HTTPRequestMethodOther}
		if method != "" {
			attrs = append(attrs, semconv.HTTPRequestMethodOriginal(method))
		}

		return "_OTHER", attrs
	}
}

func urlAttributes(requestURL *url.URL) []attribute.KeyValue {
	if requestURL == nil {
		return nil
	}

	safeURL := *requestURL
	safeURL.User = nil
	safeURL.RawQuery = ""
	safeURL.Fragment = ""

	attrs := []attribute.KeyValue{
		semconv.URLFull(safeURL.String()),
		semconv.URLScheme(safeURL.Scheme),
	}

	if safeURL.Path != "" {
		attrs = append(attrs, semconv.URLPath(safeURL.EscapedPath()))
	}

	if safeURL.Hostname() != "" {
		attrs = append(attrs, semconv.ServerAddress(safeURL.Hostname()))
	}

	if port := safeURL.Port(); port != "" {
		portNumber, err := strconv.Atoi(port)
		if err == nil {
			attrs = append(attrs, semconv.ServerPort(portNumber))
		}
	}

	return attrs
}
