package serverinterceptors

import (
	"context"
	"strconv"
	"time"

	"gitlab.deepwisdomai.com/infra/go-zero/core/metric"
	"gitlab.deepwisdomai.com/infra/go-zero/core/prometheus"
	"gitlab.deepwisdomai.com/infra/go-zero/core/timex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

const serverNamespace = "rpc_server"

var (
	metricServerReqDur = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "rpc server requests duration(ms).",
		Labels:    []string{"method"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})

	metricServerReqCodeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "rpc server requests code count.",
		Labels:    []string{"method", "code"},
	})
)

// UnaryPrometheusInterceptor returns a func that reports to the prometheus server.
func UnaryPrometheusInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
		interface{}, error) {
		if !prometheus.Enabled() {
			return handler(ctx, req)
		}

		startTime := timex.Now()
		resp, err := handler(ctx, req)
		metricServerReqDur.Observe(int64(timex.Since(startTime)/time.Millisecond), info.FullMethod)
		metricServerReqCodeTotal.Inc(info.FullMethod, strconv.Itoa(int(status.Code(err))))
		return resp, err
	}
}
