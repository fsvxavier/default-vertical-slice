package datadog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/valyala/fasthttp"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/mocktracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"gopkg.in/DataDog/dd-trace-go.v1/profiler"
)

var (
	ddAgentHost string
	started     bool
	isMock      bool
)

type DatadogAgent struct{}

func init() {
	ddAgentHost = os.Getenv("DD_ENABLED")
	if ddAgentHost == "" {
		ddAgentHost = "127.0.0.1"
		isMock = true
	}
}

func Enabled() bool {
	return len(ddAgentHost) > 0 && started
}

// StartTracing starts datadog tracing engine
// It depends on DD_ENABLED environment variable to start as read-prod env.
func StartTracing(opts ...tracer.StartOption) {
	defer func() {
		started = true
	}()
	if isMock {
		mocktracer.Start()
	}

	opts = append(opts,
		[]tracer.StartOption{
			tracer.WithLogStartup(false),
			tracer.WithEnv(os.Getenv("DD_ENV")),
			tracer.WithService(os.Getenv("DD_SERVICE")),
			tracer.WithServiceVersion(os.Getenv("DD_VERSION")),
			tracer.WithTraceEnabled(true),
			tracer.WithRuntimeMetrics(),
		}...)

	tracer.Start(opts...)
}

// StopTracing stops datadog tracing engine.
func StopTracing() {
	tracer.Stop()
	StopProfiling()
	started = false
}

// StartProfiling starts datadog tracing engine
// It depends on DD_AGENT_HOST environment variable to start as read-prod env.
func StartProfiling(opts ...profiler.Option) error {
	if started {
		if isMock {
			return nil
		}

		opts = append(opts,
			profiler.WithEnv(os.Getenv("DD_ENV")),
			profiler.WithService(os.Getenv("DD_SERVICE")),
			profiler.WithVersion(os.Getenv("DD_VERSION")),
			profiler.WithProfileTypes(
				profiler.BlockProfile,
				profiler.GoroutineProfile,
				profiler.CPUProfile,
				profiler.HeapProfile,
			),
		)

		err := profiler.Start(opts...)
		if err != nil {
			return err
		}
	}
	return nil
}

// StopTracing stops datadog tracing engine.
func StopProfiling() {
	if started {
		profiler.Stop()
	}
}

func ExtractContext(ctx context.Context) (ddtrace.SpanContext, error) {
	var traceID uint64 = 0
	var spanID uint64 = 0

	if tid := ctx.Value("x-datadog-trace-id"); tid != nil {
		if idx, ok := tid.(uint64); ok {
			traceID = idx
		}
	}

	if sid := ctx.Value("x-datadog-span-id"); sid != nil {
		if idx, ok := sid.(uint64); ok {
			traceID = idx
		}
	}

	return ExtractContextFromParent(traceID, spanID)
}

func ExtractContextFromParent(traceID, spanID uint64) (ddtrace.SpanContext, error) {
	traceData := map[string]string{
		"x-datadog-trace-id":  fmt.Sprintf("%d", traceID),
		"x-datadog-span-id":   fmt.Sprintf("%d", spanID),
		"x-datadog-parent-id": fmt.Sprintf("%d", traceID),
	}

	return tracer.Extract(tracer.TextMapCarrier(traceData))
}

func startMainSpan(operation string, startOptions ...ddtrace.StartSpanOption) *Span {
	if startOptions == nil {
		startOptions = []ddtrace.StartSpanOption{}
	}

	startOptions = append(startOptions, tracer.Measured(), tracer.StartTime(time.Now()))
	t := tracer.StartSpan(operation, startOptions...)

	return &Span{
		spanType:     SpanTypeReq,
		traceId:      t.Context().TraceID(),
		originalSpan: t,
	}
}

// StartMainReqSpan Starts a main span with Req Span properties.
func StartMainReqSpan(req *fasthttp.Request) *Span {
	startOptions := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeWeb),
		tracer.Tag(ext.HTTPMethod, string(req.Header.Method())),
		tracer.Tag(ext.HTTPURL, string(req.URI().Path())),
	}

	if len(req.Host()) != 0 {
		startOptions = append([]ddtrace.StartSpanOption{
			tracer.Tag("http.host", req.Host()),
		}, startOptions...)
	}

	return startMainSpan(string(req.URI().Path()), startOptions...)
}

// StartMainQueueSpan Starts a main span with Queue Span Properties.
func StartMainQueueSpan(queueName string) *Span {
	startOptions := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeMessageConsumer),
	}

	return startMainSpan(queueName, startOptions...)
}

// SpanChildFromParentId Starts a child span from parent traceId.
func SpanChildFromParentTraceId(traceId uint64, resourceName, spanType string) (*Span, error) {
	sc, scErr := ExtractContextFromParent(traceId, 0)

	if scErr != nil {
		return nil, scErr
	}
	return spanChild(sc, traceId, resourceName, spanType), nil
}

// SpanChildFromContext Starts a child span from parent traceId.
func SpanChildFromContext(ctx context.Context, resourceName, spanType string) (*Span, error) {
	sc, scErr := ExtractContext(ctx)

	if scErr != nil {
		return nil, scErr
	}
	return spanChild(sc, sc.TraceID(), resourceName, spanType), nil
}

// StartSpanFromContext Starts a span from context and resourceName.
func StartSpanFromContext(ctx context.Context, resourceName, spanType string) (ddtrace.Span, context.Context) {
	return tracer.StartSpanFromContext(ctx, resourceName, tracer.SpanType(spanType))
}

// SpanFromContext Creates a span from contet.
func SpanFromContext(ctx context.Context) (*Span, error) {
	span, found := tracer.SpanFromContext(ctx)

	if !found {
		return nil, errors.New("span metadata not found")
	}

	return &Span{
		spanType:     SpanTypeReq,
		traceId:      span.Context().TraceID(),
		originalSpan: span,
	}, nil
}

type SpanType string

const (
	SpanTypeReq   = ext.SpanTypeWeb
	SpanTypeQueue = ext.SpanTypeMessageConsumer
	SpanTypeDB    = ext.AppTypeDB
	SpanTypeCache = ext.AppTypeCache
)

type Span struct {
	originalSpan tracer.Span
	spanType     string
	traceId      uint64
}

func (s *Span) TraceID() uint64 {
	return s.traceId
}

func (s *Span) Context(ctx context.Context) ddtrace.SpanContext {
	return s.originalSpan.Context()
}

func spanChild(ctx ddtrace.SpanContext, traceId uint64, resourceName, spanType string) *Span {
	childStartOpts := []ddtrace.StartSpanOption{
		tracer.ResourceName(resourceName),
		tracer.ChildOf(ctx),
		tracer.SpanType(spanType),
		tracer.Measured(),
		tracer.StartTime(time.Now()),
		tracer.Tag("Trace-Id", traceId),
		tracer.Tag("trace_id", traceId),
	}

	operation := spanType
	if spanType == "" {
		operation = "http"
	}

	return &Span{
		spanType:     spanType,
		originalSpan: tracer.StartSpan(operation, childStartOpts...),
	}
}

// SpanChild Spans a child span for a sub-task in the same process.
func (s *Span) SpanChild(resourceName, spanType string) *Span {
	return spanChild(s.originalSpan.Context(), s.traceId, resourceName, spanType)
}

func (s *Span) finish(finishOptions ...tracer.FinishOption) {
	if finishOptions == nil {
		finishOptions = []tracer.FinishOption{}
	}

	finishOptions = append(finishOptions, tracer.FinishTime(time.Now()))

	s.originalSpan.Finish(finishOptions...)
}

// Finish Finishes a span.
func (s *Span) Finish() {
	s.finish()
}
