package opencensus

import (
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"

	"github.com/go-kit/kit/log"
)

// TracerOption allows for functional options to our OpenCensus tracing
// middleware.
type TracerOption func(o *TracerOptions)

// WithTracerConfig sets all configuration options at once.
func WithTracerConfig(options TracerOptions) TracerOption {
	return func(o *TracerOptions) {
		*o = options
	}
}

// WithSampler sets the sampler to use by our OpenCensus Tracer.
func WithSampler(sampler trace.Sampler) TracerOption {
	return func(o *TracerOptions) {
		o.sampler = sampler
	}
}

// WithName sets the name for an instrumented transport endpoint. If name is omitted
// at tracing middleware creation, the method of the transport or transport rpc
// name is used.
func WithName(name string) TracerOption {
	return func(o *TracerOptions) {
		o.name = name
	}
}

// WithLogger adds a Go kit logger to our OpenCensus Middleware to log SpanContext
// extract / inject errors if they occur. Default is Noop.
func WithLogger(logger log.Logger) TracerOption {
	return func(o *TracerOptions) {
		if logger != nil {
			o.logger = logger
		}
	}
}

// IsPublic ....
// func IsPublic(isPublic bool) TracerOption {
// 	return func(o *tracerOptions) {
// 		o.public = isPublic
// 	}
// }

// TracerOptions holds configuration for our tracing middlewares
type TracerOptions struct {
	sampler trace.Sampler
	name    string
	logger  log.Logger
	// public    bool
	propagate propagation.HTTPFormat
}
