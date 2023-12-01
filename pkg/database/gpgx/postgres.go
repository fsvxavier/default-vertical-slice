package gpgx

import (
	"context"
	"crypto/tls"
	"os"
	"strings"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pgInstances map[string]*PgConnection

func init() {
	pgInstances = make(map[string]*PgConnection)
}

type PgxConenector interface {
	Connect(connString string) error
}

type PgConnection struct {
	conn                  *pgxpool.Pool
	QueryExecutor         *SimpleQueryExecutor
	connString            string
	maxConns              int32
	minConns              int32
	maxConnLifetime       time.Duration
	maxConnIdletime       time.Duration
	datadogEnabled        bool
	multiTenantEnabled    bool
	multiTenantRepEnabled bool
	queryTracerEnabled    bool
}

func NewPgConnection() *PgConnection {
	pg := &PgConnection{
		maxConns:        40,
		minConns:        20,
		maxConnLifetime: time.Second * 9,
		maxConnIdletime: time.Second * 3,
	}
	pgInstances["main"] = pg
	return pg
}

func (pgc *PgConnection) Pool() *pgxpool.Pool {
	return pgc.conn
}

func (pgc *PgConnection) SetDatadogEnable(enabled bool) *PgConnection {
	pgc.datadogEnabled = enabled
	return pgc
}

func (pgc *PgConnection) SetMultiTenantEnabled(enabled bool) *PgConnection {
	pgc.multiTenantEnabled = enabled
	return pgc
}

func (pgc *PgConnection) SetMultiTenantRepEnabled(enabled bool) *PgConnection {
	pgc.multiTenantRepEnabled = enabled
	return pgc
}

func (pgc *PgConnection) SetQueryTracerEnabled(enabled bool) *PgConnection {
	pgc.queryTracerEnabled = enabled
	return pgc
}

func (pgc *PgConnection) SetMaxConns(vnumber int32) *PgConnection {
	pgc.maxConns = vnumber
	return pgc
}

func (pgc *PgConnection) SetMinConns(vnumber int32) *PgConnection {
	pgc.minConns = vnumber
	return pgc
}

func (pgc *PgConnection) SetMaxConnLifetime(vtime time.Duration) *PgConnection {
	pgc.maxConnLifetime = vtime
	return pgc
}

func (pgc *PgConnection) SetMaxConnIdleTime(vtime time.Duration) *PgConnection {
	pgc.maxConnIdletime = vtime
	return pgc
}

// NewPool creates a new Pool and immediately establishes one connection.
// maxConns is the maximum size of the pool. The default is the max(4, runtime.NumCPU()).
func (pgc *PgConnection) NewPool(ctx context.Context, connString string) error {
	pgc.connString = connString

	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	config.ConnConfig.TLSConfig = tlsConfig

	config.ConnConfig.RuntimeParams["timezone"] = "UTC"
	if os.Getenv("DB_QUERY_MODE_EXEC") == "SIMPLE_PROTOCOL" {
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	} else {
		config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	}

	// Creates a new pool with the given configuration.
	// MaxConns is the maximum size of the pool. The default is the greater of 4 or runtime.NumCPU().
	config.MaxConns = pgc.maxConns
	config.MinConns = pgc.minConns
	config.MaxConnLifetime = pgc.maxConnLifetime
	config.MaxConnIdleTime = pgc.maxConnIdletime

	config.ConnConfig.Tracer = &TracerConfig{
		QueryTracerEnabled: pgc.isQueryTracerEnabled(),
		DatadogEnabled:     pgc.isDatadogEnabled(),
	}

	if pgc.multiTenantEnabled && !pgc.multiTenantRepEnabled {
		mtc := &MultiTenantConfig{}
		// BeforeAcquire is called before a connection is acquired from the pool.
		// It must return true to allow the acquisition or false to indicate that the connection should be destroyed and
		// a different connection should be acquired.
		config.BeforeAcquire = mtc.beforeAcquireHook

		// AfterRelease is called after a connection is released, but before it is returned to the pool.
		// It must return true to return the connection to the pool or false to destroy the connection.
		config.AfterRelease = mtc.afterReleaseHook
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}
	pgc.conn = pool

	return nil
}

func (pgc *PgConnection) isDatadogEnabled() bool {
	return pgc.datadogEnabled && os.Getenv("DD_AGENT_HOST") != ""
}

func (pgc *PgConnection) isQueryTracerEnabled() bool {
	return pgc.queryTracerEnabled
}

func (pgc *PgConnection) Reconnect(ctx context.Context) error {
	return pgc.NewPool(ctx, pgc.connString)
}

func (pgc *PgConnection) Close() {
	if pgc.Pool() != nil {
		pgc.Pool().Close()
	}
}

func (pgc *PgConnection) Stat() *pgxpool.Stat {
	if pgc.Pool() != nil {
		return pgc.Pool().Stat()
	}
	return nil
}

func (pgc *PgConnection) queryFor(ctx context.Context, tx *pgx.Tx, dst any, many bool, sql string, arguments ...any) error {
	qr, err := pgc.query(ctx, tx, sql, arguments...)
	if err != nil {
		if _, ok := err.(*NotConnectedError); ok {
			return err
		}
		return NewPgError(err.Error())
	}

	defer qr.Close()

	if dst != nil {
		if many {
			err = pgxscan.ScanAll(dst, qr)
		} else {
			err = pgxscan.ScanOne(dst, qr)
		}
	}

	if err != nil {
		return NewPgError(err.Error())
	}

	return nil
}

func (pgc *PgConnection) query(ctx context.Context, tx *pgx.Tx, sql string, arguments ...any) (pgx.Rows, error) {
	if pgc.conn == nil {
		return nil, new(NotConnectedError)
	}
	f := pgc.conn.Query

	if tx != nil {
		f = (*tx).Query
	}

	if ctx == nil {
		ctx = context.TODO()
	}
	r, err := f(ctx, sql, arguments...)
	if err != nil {
		if strings.HasPrefix(err.Error(), "failed to connect") {
			return nil, new(NotConnectedError)
		}
		return nil, err
	}

	rErr := r.Err()
	if rErr != nil {
		return nil, rErr
	}

	return r, nil
}

func Pg(name ...string) *PgConnection {
	if len(name) == 0 {
		name = append(name, "main")
	}

	n := name[0]
	if _, ok := pgInstances[n]; !ok {
		pgInstances[n] = &PgConnection{}
		pgInstances[n].QueryExecutor = &SimpleQueryExecutor{pgConn: pgInstances[n]}
	}

	return pgInstances[n]
}
