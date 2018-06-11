package opencensus

import (
	"context"

	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	kitgrpc "github.com/go-kit/kit/transport/grpc"
)

const propagationKey = "grpc-trace-bin"

// GRPCClientTrace enables OpenCensus tracing of a Go kit gRPC transport client.
func GRPCClientTrace(options ...TracerOption) kitgrpc.ClientOption {
	config := TracerOptions{
		name:    "",
		sampler: trace.AlwaysSample(),
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
			return trace.NewContext(ctx, span)
		},
	)

	clientFinalizer := kitgrpc.ClientFinalizer(
		func(ctx context.Context, err error) {
			if span := trace.FromContext(ctx); span != nil {
				s, ok := status.FromError(err)
				if ok {
					span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
				} else {
					span.SetStatus(trace.Status{Code: int32(codes.Unknown), Message: err.Error()})
				}
				span.End()
			}
		},
	)

	return func(c *kitgrpc.Client) {
		clientBefore(c)
		clientFinalizer(c)
	}

}

// GRPCServerTrace enables OpenCensus tracing of a Go kit gRPC transport server.
func GRPCServerTrace(options ...TracerOption) kitgrpc.ServerOption {
	config := TracerOptions{}

	for _, option := range options {
		option(&config)
	}

	serverBefore := kitgrpc.ServerBefore(
		func(ctx context.Context, md metadata.MD) context.Context {
			var (
				ok   bool
				name string
			)

			if config.name != "" {
				name = config.name
			} else {
				name, ok = ctx.Value(kitgrpc.ContextKeyRequestMethod).(string)
				if !ok || name == "" {
					// we can't find the gRPC method. probably the
					// unaryInterceptor was not wired up.
					name = "unknown grpc method"
				}
			}

			traceContext := md[propagationKey]

			if len(traceContext) > 0 {
				traceContextBinary := []byte(traceContext[0])
				spanContext, ok := propagation.FromBinary(traceContextBinary)
				if ok {
					ctx, _ = trace.StartSpanWithRemoteParent(
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
			if span := trace.FromContext(ctx); span != nil {
				s, ok := status.FromError(err)
				if ok {
					span.SetStatus(trace.Status{Code: int32(s.Code()), Message: s.Message()})
				} else {
					span.SetStatus(trace.Status{Code: int32(codes.Internal), Message: err.Error()})
				}
				span.End()
			}
		},
	)

	return func(s *kitgrpc.Server) {
		serverBefore(s)
		serverFinalizer(s)
	}
}
