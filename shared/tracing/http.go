package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// WrapHandlerFunc wraps an http.HandlerFunc with OpenTelemetry instrumentation.
// The operation name appears as the span name in Jaeger.
func WrapHandlerFunc(handler http.HandlerFunc, operation string) http.Handler {
	return otelhttp.NewHandler(handler, operation)
}
