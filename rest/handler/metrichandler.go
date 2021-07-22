package handler

import (
	"net/http"

	"gitlab.deepwisdomai.com/infra/go-zero/core/stat"
	"gitlab.deepwisdomai.com/infra/go-zero/core/timex"
)

// MetricHandler returns a middleware that stat the metrics.
func MetricHandler(metrics *stat.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := timex.Now()
			defer func() {
				metrics.Add(stat.Task{
					Duration: timex.Since(startTime),
				})
			}()

			next.ServeHTTP(w, r)
		})
	}
}
