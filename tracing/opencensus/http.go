package opencensus

import (
	"context"
	"net/http"

	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"

	kithttp "github.com/go-kit/kit/transport/http"
)

// HTTPClientTrace enables native OpenCensus tracing of a Go kit HTTP transport
// Client.
func HTTPClientTrace(options ...TracerOption) kithttp.ClientOption {
	config := tracerOptions{
		name: "",
		// public:    true,
		sampler:   trace.AlwaysSample(),
		propagate: &b3.HTTPFormat{},
	}

	for _, option := range options {
		option(&config)
	}

	clientBefore := kithttp.ClientBefore(
		func(ctx context.Context, req *http.Request) context.Context {
			var name string

			if config.name != "" {
				name = config.name
			} else {
				// OpenCensus states Path being default naming for a client span
				name = req.Method + " " + req.URL.Path
			}

			span := trace.NewSpan(
				name,
				trace.FromContext(ctx),
				trace.StartOptions{
					Sampler:  config.sampler,
					SpanKind: trace.SpanKindClient,
				},
			)

			span.AddAttributes(
				trace.StringAttribute(ochttp.HostAttribute, req.URL.Host),
				trace.StringAttribute(ochttp.MethodAttribute, req.Method),
				trace.StringAttribute(ochttp.PathAttribute, req.URL.Path),
				trace.StringAttribute(ochttp.UserAgentAttribute, req.UserAgent()),
			)

			if config.propagate != nil {
				config.propagate.SpanContextToRequest(span.SpanContext(), req)
			}

			return trace.NewContext(ctx, span)
		},
	)

	clientAfter := kithttp.ClientAfter(
		func(ctx context.Context, res *http.Response) context.Context {
			if span := trace.FromContext(ctx); span != nil {
				span.SetStatus(ochttp.TraceStatus(res.StatusCode, http.StatusText(res.StatusCode)))
				span.AddAttributes(
					trace.Int64Attribute(ochttp.StatusCodeAttribute, int64(res.StatusCode)),
				)
			}
			return ctx
		},
	)

	clientFinalizer := kithttp.ClientFinalizer(
		func(ctx context.Context, err error) {
			if span := trace.FromContext(ctx); span != nil {
				if err != nil {
					span.SetStatus(trace.Status{Code: 2, Message: err.Error()})
				}
				span.End()
			}
		},
	)

	return func(c *kithttp.Client) {
		clientBefore(c)
		clientAfter(c)
		clientFinalizer(c)
	}
}

// HTTPServerTrace enables native OpenCensus tracing of a Go kit HTTP transport
// Server.
func HTTPServerTrace(options ...TracerOption) kithttp.ServerOption {
	config := tracerOptions{
		name: "",
		//public:    true,
		sampler:   trace.AlwaysSample(),
		propagate: &b3.HTTPFormat{},
	}

	for _, option := range options {
		option(&config)
	}

	serverBefore := kithttp.ServerBefore(
		func(ctx context.Context, req *http.Request) context.Context {
			var (
				spanContext trace.SpanContext
				span        *trace.Span
				name        string
			)

			if config.name != "" {
				name = config.name
			} else {
				name = req.Method + " " + req.URL.Path
			}

			if config.propagate != nil {
				spanContext, _ = config.propagate.SpanContextFromRequest(req)
			}

			ctx, span = trace.StartSpanWithRemoteParent(
				ctx,
				name,
				spanContext,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithSampler(config.sampler),
			)

			span.AddAttributes(
				trace.StringAttribute(ochttp.MethodAttribute, req.Method),
				trace.StringAttribute(ochttp.PathAttribute, req.URL.Path),
			)

			return ctx
		},
	)

	serverFinalizer := kithttp.ServerFinalizer(
		func(ctx context.Context, code int, r *http.Request) {
			if span := trace.FromContext(ctx); span != nil {
				span.SetStatus(ochttp.TraceStatus(code, http.StatusText(code)))

				if rs, ok := ctx.Value(kithttp.ContextKeyResponseSize).(int64); ok {
					span.AddAttributes(
						trace.Int64Attribute("http.response_size", rs),
					)
				}

				span.End()
			}
		},
	)

	return func(s *kithttp.Server) {
		serverBefore(s)
		serverFinalizer(s)
	}
}
