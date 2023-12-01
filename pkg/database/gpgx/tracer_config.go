package gpgx

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	spanOperation  = "pgx.query"
	queryTypeQuery = "Query"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TracerConfig contains configs to tracing and would implement QueryTracer, BatchTracer,
// CopyFromTracer, PrepareTracer and ConnectTracer from pgx.
type TracerConfig struct {
	spanMap            map[string]*tracer.Span
	DatadogEnabled     bool
	QueryTracerEnabled bool
}

type spanMapKeyType string

var (
	spanMapKey spanMapKeyType = "spanMapKey"
	mtx        sync.Mutex
)

type Option func(*TracerConfig)

// TraceQueryStart is called at the beginning of Query, QueryRow, and Exec calls. The returned context is used for the
// rest of the call and will be passed to TraceQueryEnd.
func (cfg *TracerConfig) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	// if strings.Index(data.SQL, "set_config") > 0 {
	// 	return ctx
	// }

	if cfg.QueryTracerEnabled {
		dataArgs, err := json.Marshal(data.Args)
		if err != nil {
			fmt.Println("Error marshal query args")
		}
		fmt.Printf("\n\nQUERY: %s - ARGS: - %s\n\n", data.SQL, string(dataArgs))
	}

	if cfg.DatadogEnabled {
		opts := []ddtrace.StartSpanOption{
			tracer.SpanType(ext.SpanTypeSQL),
			tracer.Tag(ext.Component, "jackc/pgx"),
			tracer.Tag(ext.ResourceName, data.SQL),
			tracer.Tag("db.system", "postgresql"),
			tracer.Tag("sql.query_type", queryTypeQuery),
			tracer.Tag("dd.env", os.Getenv("DD_ENV")),
			tracer.Tag("dd.version", os.Getenv("DD_VERSION")),
		}

		sn := os.Getenv("DD_SERVICE_DB")
		if sn == "" {
			sn = os.Getenv("DD_SERVICE_DB")
		}
		if sn != "" {
			opts = append(opts, tracer.ServiceName(sn))
		} else {
			opts = append(opts, tracer.ServiceName(os.Getenv("DD_SERVICE")+".db"))
		}

		span, _ := tracer.StartSpanFromContext(ctx, spanOperation, opts...)
		mtx.Lock()
		if cfg.spanMap == nil {
			cfg.spanMap = make(map[string]*ddtrace.Span)
		}

		uuidKey := uuid.New().String()
		cfg.spanMap[uuidKey] = &span
		mtx.Unlock()

		return context.WithValue(ctx, spanMapKey, uuidKey)
	}

	return ctx
}

// TraceQueryEnd traces the end of the query, implementing pgx.QueryTracer.
func (cfg *TracerConfig) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	uuidKey := ctx.Value(spanMapKey)
	if uuidKey != nil {
		if k, ok := uuidKey.(string); ok {
			mtx.Lock()
			span := cfg.spanMap[k]
			(*span).Finish()

			delete(cfg.spanMap, k)
			mtx.Unlock()
		}
	}
}
