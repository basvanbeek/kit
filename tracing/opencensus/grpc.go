package opencensus

import (
	"context"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/go-kit/kit/log"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

const propagationKey = "grpc-trace-bin"

// GRPCClientTrace enables native OpenCensus tracing of a Go kit gRPC transport
// Client.
func GRPCClientTrace(options ...TracerOption) kitgrpc.ClientOption {
	config := tracerOptions{
		name:    "",
		sampler: trace.AlwaysSample(),
		// logger:  log.NewNopLogger(),
	}

	for _, option := range options {
		option(&config)
	}

	clientBefore := kitgrpc.ClientBefore(
		func(ctx context.Context, md *metadata.MD) context.Context {
			var name string

			if config.name != "" {
				name = config.name
			} else {
				name = ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
			}

			span := trace.NewSpan(
				name,
				trace.FromContext(ctx),
				trace.StartOptions{
					Sampler:  config.sampler,
					SpanKind: trace.SpanKindClient,
				},
			)

			traceContextBinary := string(propagation.Binary(span.SpanContext()))
			(*md)[propagationKey] = append((*md)[propagationKey], traceContextBinary)
			return ctx
		},
	)

	clientFinalizer := kitgrpc.ClientFinalizer(
		func(ctx context.Context, err error) {
			span := trace.FromContext(ctx)

			s, ok := status.FromError(err)
			if ok {
				span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
			} else {
				span.SetStatus(trace.Status{Code: int32(codes.Internal), Message: err.Error()})
			}
		},
	)

	return func(c *kitgrpc.Client) {
		clientBefore(c)
		clientFinalizer(c)
	}

}

// GRPCServerTrace enables native Zipkin tracing of a Go kit gRPC transport
// Server.
func GRPCServerTrace(options ...TracerOption) kitgrpc.ServerOption {
	config := tracerOptions{
		name:   "",
		logger: log.NewNopLogger(),
	}

	for _, option := range options {
		option(&config)
	}

	serverBefore := kitgrpc.ServerBefore(
		func(ctx context.Context, md metadata.MD) context.Context {
			var (
				name string
			)

			rpcMethod, ok := ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
			if !ok {
				config.logger.Log("unable to retrieve method name: missing gRPC interceptor hook")
			}

			if config.name != "" {
				name = config.name
			} else {
				name = rpcMethod
			}

			traceContext := md[propagationKey]

			if len(traceContext) > 0 {
				traceContextBinary := []byte(traceContext[0])
				spanContext, ok := propagation.FromBinary(traceContextBinary)
				if ok {
					ctx, _ := trace.StartSpanWithRemoteParent(
						ctx,
						name,
						spanContext,
						trace.WithSpanKind(trace.SpanKindServer),
						trace.WithSampler(config.sampler),
					)
					return ctx
				}
			}
			ctx, _ = trace.StartSpan(
				ctx,
				name,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithSampler(config.sampler),
			)

			return ctx
		},
	)

	serverFinalizer := kitgrpc.ServerFinalizer(
		func(ctx context.Context, err error) {
			span := trace.FromContext(ctx)

			s, ok := status.FromError(err)
			if ok {
				span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
			} else {
				span.SetStatus(trace.Status{Code: int32(codes.Internal), Message: err.Error()})
			}
		},
	)

	return func(s *kitgrpc.Server) {
		serverBefore(s)
		serverFinalizer(s)
	}
}
