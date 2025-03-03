package serverinterceptors

import (
	"context"

	"gitlab.deepwisdomai.com/infra/go-zero/core/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UnaryTracingInterceptor returns a func that handles tracing with given service name.
func UnaryTracingInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		carrier, err := trace.Extract(trace.GrpcFormat, md)
		if err != nil {
			return handler(ctx, req)
		}

		ctx, span := trace.StartServerSpan(ctx, carrier, serviceName, info.FullMethod)
		defer span.Finish()
		return handler(ctx, req)
	}
}
