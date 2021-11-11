package trace

import (
	"net/http"

	"gitlab.deepwisdomai.com/infra/go-zero/rest"
	"google.golang.org/grpc/metadata"
)

func InjectTracing() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			md := metadata.Pairs(
				"x-request-id", r.Header.Get("X-Request-Id"),
				"X-B3-TraceId", r.Header.Get("X-B3-Traceid"),
				"X-B3-SpanId", r.Header.Get("X-B3-Spanid"),
				"X-B3-TraceId", r.Header.Get("X-B3-TraceId"),
				"x-B3-Sampled", r.Header.Get("X-B3-Sampled"),
				"x-B3-Flags", r.Header.Get("x-b3-flags"),
				"x-Ot-Span-Context", r.Header.Get("x-ot-span-context"),
			)
			r.WithContext(metadata.NewIncomingContext(r.Context(), md))
			next(w, r)
		}
	}
}
