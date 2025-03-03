package internal

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.deepwisdomai.com/infra/go-zero/zrpc/internal/balancer/p2c"
	"gitlab.deepwisdomai.com/infra/go-zero/zrpc/internal/clientinterceptors"
	"gitlab.deepwisdomai.com/infra/go-zero/zrpc/internal/resolver"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	ot "github.com/opentracing/opentracing-go"
)

const (
	dialTimeout = time.Second * 3
	separator   = '/'
)

func init() {
	resolver.RegisterResolver()
}

type (
	// Client interface wraps the Conn method.
	Client interface {
		Conn() *grpc.ClientConn
	}

	// A ClientOptions is a client options.
	ClientOptions struct {
		Timeout     time.Duration
		DialOptions []grpc.DialOption
	}

	// ClientOption defines the method to customize a ClientOptions.
	ClientOption func(options *ClientOptions)

	client struct {
		conn *grpc.ClientConn
	}
)

// NewClient returns a Client.
func NewClient(target string, opts ...ClientOption) (Client, error) {
	var cli client
	opts = append([]ClientOption{WithDialOption(grpc.WithBalancerName(p2c.Name))}, opts...)
	if err := cli.dial(target, opts...); err != nil {
		return nil, err
	}

	return &cli, nil
}

func (c *client) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *client) buildDialOptions(opts ...ClientOption) []grpc.DialOption {
	var cliOpts ClientOptions
	for _, opt := range opts {
		opt(&cliOpts)
	}

	options := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		WithUnaryClientInterceptors(
			clientinterceptors.TracingInterceptor,
			clientinterceptors.DurationInterceptor,
			clientinterceptors.BreakerInterceptor,
			clientinterceptors.PrometheusInterceptor,
			clientinterceptors.TimeoutInterceptor(cliOpts.Timeout),
		),
	}

	// tracing
	options = append(options, grpc.WithStreamInterceptor(
		grpc_opentracing.StreamClientInterceptor(
			grpc_opentracing.WithTracer(ot.GlobalTracer()))))
	options = append(options, grpc.WithUnaryInterceptor(
		grpc_opentracing.UnaryClientInterceptor(
			grpc_opentracing.WithTracer(ot.GlobalTracer()))))

	return append(options, cliOpts.DialOptions...)
}

func (c *client) dial(server string, opts ...ClientOption) error {
	options := c.buildDialOptions(opts...)
	timeCtx, cancel := context.WithTimeout(context.Background(), dialTimeout)
	defer cancel()
	conn, err := grpc.DialContext(timeCtx, server, options...)
	if err != nil {
		service := server
		if errors.Is(err, context.DeadlineExceeded) {
			pos := strings.LastIndexByte(server, separator)
			// len(server) - 1 is the index of last char
			if 0 < pos && pos < len(server)-1 {
				service = server[pos+1:]
			}
		}
		return fmt.Errorf("rpc dial: %s, error: %s, make sure rpc service %q is already started",
			server, err.Error(), service)
	}

	c.conn = conn
	return nil
}

// WithDialOption returns a func to customize a ClientOptions with given dial option.
func WithDialOption(opt grpc.DialOption) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, opt)
	}
}

// WithTimeout returns a func to customize a ClientOptions with given timeout.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(options *ClientOptions) {
		options.Timeout = timeout
	}
}

// WithUnaryClientInterceptor returns a func to customize a ClientOptions with given interceptor.
func WithUnaryClientInterceptor(interceptor grpc.UnaryClientInterceptor) ClientOption {
	return func(options *ClientOptions) {
		options.DialOptions = append(options.DialOptions, WithUnaryClientInterceptors(interceptor))
	}
}
