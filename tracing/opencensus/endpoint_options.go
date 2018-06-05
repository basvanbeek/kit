package opencensus

import "go.opencensus.io/trace"

// EndpointOptions holds the options for tracing an endpoint
type EndpointOptions struct {
	IgnoreBusinessError bool
	Attributes          []trace.Attribute
}

// EndpointOption allows for functional options to our OpenCensus endpoint
// tracing middleware.
type EndpointOption func(*EndpointOptions)

// WithEndpointConfig sets all configuration options at once.
func WithEndpointConfig(options EndpointOptions) EndpointOption {
	return func(o *EndpointOptions) {
		*o = options
	}
}

// WithEndpointAttributes sets the default attributes for the spans created by
// the Endpoint tracer.
func WithEndpointAttributes(attrs ...trace.Attribute) EndpointOption {
	return func(o *EndpointOptions) {
		o.Attributes = attrs
	}
}

// WithIgnoreBusinessError will not treat a business error identified through
// the Failer interface as a span error.
func WithIgnoreBusinessError(val bool) EndpointOption {
	return func(o *EndpointOptions) {
		o.IgnoreBusinessError = val
	}
}
