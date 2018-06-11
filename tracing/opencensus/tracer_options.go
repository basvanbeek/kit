package opencensus

import (
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"

	"github.com/go-kit/kit/log"
)

// defaultHTTPPropagate holds OpenCensus' default HTTP propagation format which
// currently is Zipkin's B3.
var defaultHTTPPropagate propagation.HTTPFormat = &b3.HTTPFormat{}

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

// IsPublic should be set to true for publicly accessible servers and for
// clients that should not propagate their current trace metadata.
// On the server side a new trace will always be started regardless of any
// trace metadata being found in the incoming request. If any trace metadata
// is found, it will be added as a linked trace instead.
func IsPublic(isPublic bool) TracerOption {
	return func(o *TracerOptions) {
		o.public = isPublic
	}
}

// WithHTTPPropagation sets the propagation handlers for the HTTP transport
// middlewares. If used on a non HTTP transport this is a noop.
func WithHTTPPropagation(p propagation.HTTPFormat) TracerOption {
	return func(o *TracerOptions) {
		if p == nil {
			// reset to default OC HTTP format
			o.httpPropagate = defaultHTTPPropagate
		}
		o.httpPropagate = p
	}
}

// TracerOptions holds configuration for our tracing middlewares
type TracerOptions struct {
	sampler       trace.Sampler
	name          string
	logger        log.Logger
	public        bool
	httpPropagate propagation.HTTPFormat
}
