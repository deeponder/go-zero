package logx

import (
	"context"
	"fmt"
	"io"
	"time"

	"gitlab.deepwisdomai.com/infra/go-zero/core/timex"
	"gitlab.deepwisdomai.com/infra/go-zero/core/trace/tracespec"
)

type traceLogger struct {
	logEntry
	Trace string `json:"trace,omitempty"`
	Span  string `json:"span,omitempty"`
	ctx   context.Context
}

func (l *traceLogger) Error(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprint(v...), durationCallerDepth))
	}
}

func (l *traceLogger) Errorf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(errorLog, levelError, formatWithCaller(fmt.Sprintf(format, v...), durationCallerDepth))
	}
}

func (l *traceLogger) Info(v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprint(v...))
	}
}

func (l *traceLogger) Infof(format string, v ...interface{}) {
	if shouldLog(InfoLevel) {
		l.write(infoLog, levelInfo, fmt.Sprintf(format, v...))
	}
}

func (l *traceLogger) Slow(v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprint(v...))
	}
}

func (l *traceLogger) Slowf(format string, v ...interface{}) {
	if shouldLog(ErrorLevel) {
		l.write(slowLog, levelSlow, fmt.Sprintf(format, v...))
	}
}

func (l *traceLogger) WithDuration(duration time.Duration) Logger {
	l.Duration = timex.ReprOfDuration(duration)
	return l
}

func (l *traceLogger) write(writer io.Writer, level, content string) {
	l.Timestamp = getTimestamp()
	l.Level = level
	l.Content = content
	l.Trace = traceIdFromIstio(l.ctx)
	l.Span = spanIdFromIstio(l.ctx)
	outputJson(writer, l)
}

// WithContext sets ctx to log, for keeping tracing information.
func WithContext(ctx context.Context) Logger {
	return &traceLogger{
		ctx: ctx,
	}
}

func spanIdFromContext(ctx context.Context) string {
	t, ok := ctx.Value(tracespec.TracingKey).(tracespec.Trace)
	if !ok {
		return ""
	}

	return t.SpanId()
}

func traceIdFromContext(ctx context.Context) string {
	t, ok := ctx.Value(tracespec.TracingKey).(tracespec.Trace)
	if !ok {
		return ""
	}

	return t.TraceId()
}

func spanIdFromIstio(ctx context.Context) string {
	spanId, ok := ctx.Value("x-b3-traceid").(string)
	if !ok {
		return ""
	}

	return spanId
}

func traceIdFromIstio(ctx context.Context) string {
	traceId, ok := ctx.Value("x-b3-spanid").(string)
	if !ok {
		return ""
	}

	return traceId
}
