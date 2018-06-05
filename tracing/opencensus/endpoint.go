package opencensus

import (
	"context"
	"strconv"

	"go.opencensus.io/trace"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd/lb"
)

// TraceEndpointDefaultName is the default endpoint span name to use.
const TraceEndpointDefaultName = "gokit/endpoint"

// TraceEndpoint returns an Endpoint middleware, tracing a Go kit endpoint.
// This endpoint tracer should be used in combination with a Go kit Transport
// tracing middleware, generic OpenCensus transport middleware or custom before
// and after transport functions as service propagation of SpanContext is not
// provided in this middleware.
func TraceEndpoint(name string, options ...EndpointOption) endpoint.Middleware {
	if name == "" {
		name = TraceEndpointDefaultName
	}
	cfg := &EndpointOptions{}
	for _, o := range options {
		o(cfg)
	}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx, span := trace.StartSpan(ctx, name)
			if len(cfg.Attributes) > 0 {
				span.AddAttributes(cfg.Attributes...)
			}
			defer func() {
				if err != nil {
					if lberr, ok := err.(lb.RetryError); ok {
						attrs := make([]trace.Attribute, 0, len(lberr.RawErrors))
						for idx, err := range lberr.RawErrors {
							attrs = append(attrs, trace.StringAttribute(
								"gokit.retry.error."+strconv.Itoa(idx+1), err.Error(),
							))
						}
						span.AddAttributes(attrs...)
						span.SetStatus(trace.Status{
							Code:    trace.StatusCodeUnknown,
							Message: lberr.Final.Error(),
						})
					} else {
						span.SetStatus(trace.Status{
							Code:    trace.StatusCodeUnknown,
							Message: err.Error(),
						})
					}
				} else if res, ok := response.(failer); ok {
					if err = res.Failed(); err != nil {
						span.AddAttributes(
							trace.StringAttribute("gokit.business.error", err.Error()),
						)
						if cfg.IgnoreBusinessError {
							span.SetStatus(trace.Status{Code: trace.StatusCodeOK})
						} else {
							span.SetStatus(trace.Status{
								Code:    trace.StatusCodeUnknown,
								Message: err.Error(),
							})
						}
					}
				} else {
					span.SetStatus(trace.Status{Code: trace.StatusCodeOK})
				}
				span.End()
			}()
			response, err = next(ctx, request)
			return
		}
	}
}

// failer is a Go kit idiomatic interface as shown in the addsvc example.
// a typical response payload which can hold business logic er§rors should
// implement the Failed() method.
type failer interface {
	Failed() error
}
