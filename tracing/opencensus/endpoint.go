package opencensus

import (
	"context"

	"go.opencensus.io/trace"

	"github.com/go-kit/kit/endpoint"
)

// TraceEndpoint returns an Endpoint middleware, tracing a Go kit endpoint.
// This endpoint tracer should be used in combination with a Go kit Transport
// tracing middleware, generic OpenCensus transport middleware or custom before
// and after transport functions as propagation of SpanContext is not provided
// in this middleware.
func TraceEndpoint(name string) endpoint.Middleware {
	if name == "" {
		name = "kit/endpoint"
	}
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			ctx, span := trace.StartSpan(ctx, name)
			defer func() {
				if err != nil {
					span.SetStatus(trace.Status{
						trace.StatusCodeUnknown, err.Error(),
					})
				} else if res, ok := response.(failer); ok {
					if err = res.Failed(); err != nil {
						span.SetStatus(trace.Status{
							trace.StatusCodeUnknown, err.Error(),
						})
					}
				}
				span.End()
			}()
			response, err = next(ctx, request)
			return
		}
	}
}

// failer is a go kit idiomatic interface as shown in the addsvc example.
// a typical response payload which can hold business logic errors should
// implement the Failed() method.
type failer interface {
	Failed() error
}
