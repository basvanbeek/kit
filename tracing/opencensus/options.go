package opencensus

import (
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"

	"github.com/go-kit/kit/log"
)

// TracerOption allows for functional options to our OpenCensus tracing
// middleware.
type TracerOption func(o *tracerOptions)

// Sampler sets the sampler to use by our OpenCensus Tracer.
func Sampler(sampler trace.Sampler) TracerOption {
	return func(o *tracerOptions) {
		o.sampler = sampler
	}
}

// Name sets the name for an instrumented transport endpoint. If name is omitted
// at tracing middleware creation, the method of the transport or transport rpc
// name is used.
func Name(name string) TracerOption {
	return func(o *tracerOptions) {
		o.name = name
	}
}

// Logger adds a Go kit logger to our OpenCensus Middleware to log SpanContext
// extract / inject errors if they occur. Default is Noop.
func Logger(logger log.Logger) TracerOption {
	return func(o *tracerOptions) {
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

type tracerOptions struct {
	sampler trace.Sampler
	name    string
	logger  log.Logger
	// public    bool
	propagate propagation.HTTPFormat
}
