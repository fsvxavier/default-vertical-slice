package redis

import (
	"context"
	"crypto/tls"
	"time"

	rgo "github.com/gomodule/redigo/redis"
	rgoc "github.com/mna/redisc"
	redigotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gomodule/redigo"
)

type RedigoPoolOptions struct {
	Context          context.Context
	TlsConfig        *tls.Config
	Password         string
	TraceServiceName string
	ClientName       string
	Addresses        []string
	MaxConnLifetime  time.Duration
	IdleTimeout      time.Duration
	Db               int
	MaxIdle          int
	PoolSize         int
	Database         int
	MaxActive        int
	SkipVerify       bool
	UsageTLS         bool
	ExecutePing      bool
}

const (
	TRUE                = true
	FALSE               = false
	MAX_IDLE            = 300
	MAX_ACTIVE          = 1000
	CONNECT_TIMEOUT_SEC = 5
)

func NewRedigo(ctx context.Context, opt *RedigoPoolOptions) (rdbg Redigo, err error) {
	options := rdbg.prepareOptions(ctx, opt)
	if len(options.Addresses) > 1 {
		cluster, errCluster := rdbg.createCluster(&options)
		if errCluster != nil {
			return Redigo{}, errCluster
		}

		// initialize its mapping
		errRefresh := cluster.Refresh()
		if errRefresh != nil {
			return Redigo{}, errRefresh
		}

		rdbg.Cluster = cluster
	} else {
		pool, errPool := rdbg.createPool(ctx, &options)
		if errPool != nil {
			return Redigo{}, errPool
		}
		rdbg.Pool = pool
	}

	return rdbg, err
}

type Redigo struct {
	ctx              context.Context
	Pool             *rgo.Pool
	Cluster          *rgoc.Cluster
	tlsConfig        *tls.Config
	password         string
	traceServiceName string
	clientName       string
	addresses        []string
	idleTimeout      time.Duration
	maxConnLifetime  time.Duration
	maxIdle          int
	poolSize         int
	database         int
	maxActive        int
	skipVerify       bool
	usageTLS         bool
}

func (rdbg *Redigo) createCluster(opts *RedigoPoolOptions) (cluster *rgoc.Cluster, err error) {
	// create the cluster
	cluster = &rgoc.Cluster{
		StartupNodes: opts.Addresses,
		DialOptions: []rgo.DialOption{
			rgo.DialConnectTimeout(CONNECT_TIMEOUT_SEC * time.Second),
		},
		CreatePool: rdbg.creatingPool,
	}

	return cluster, nil
}

func (rdbg *Redigo) creatingPool(addr string, opts ...rgo.DialOption) (*rgo.Pool, error) {
	return &rgo.Pool{
		MaxIdle:         rdbg.maxIdle,
		MaxActive:       rdbg.maxActive,
		IdleTimeout:     rdbg.idleTimeout,
		MaxConnLifetime: rdbg.maxConnLifetime,
		DialContext: func(ctx context.Context) (conn rgo.Conn, err error) {
			connDial, errDial := redigotrace.DialContext(rdbg.ctx, "tcp", addr,
				redigotrace.WithContextConnection(),
				rgo.DialConnectTimeout(CONNECT_TIMEOUT_SEC*time.Second),
				rgo.DialUseTLS(rdbg.usageTLS),
				rgo.DialTLSConfig(rdbg.tlsConfig),
				redigotrace.WithServiceName(rdbg.traceServiceName),
				rgo.DialTLSSkipVerify(rdbg.skipVerify),
				rgo.DialDatabase(rdbg.database),
			)
			return connDial, errDial
		},
		TestOnBorrow: func(c rgo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}, nil
}

func (rdbg *Redigo) createPool(ctx context.Context, opts *RedigoPoolOptions) (pool *rgo.Pool, err error) {
	pool = &rgo.Pool{
		MaxIdle:         rdbg.maxIdle,
		MaxActive:       rdbg.maxActive,
		IdleTimeout:     rdbg.idleTimeout,
		MaxConnLifetime: rdbg.maxConnLifetime,
		DialContext: func(ctx context.Context) (conn rgo.Conn, err error) {
			connDial, errDial := redigotrace.DialContext(rdbg.ctx, "tcp", opts.Addresses[0],
				redigotrace.WithContextConnection(),
				rgo.DialConnectTimeout(CONNECT_TIMEOUT_SEC*time.Second),
				rgo.DialUseTLS(rdbg.usageTLS),
				rgo.DialTLSConfig(rdbg.tlsConfig),
				redigotrace.WithServiceName(rdbg.traceServiceName),
				rgo.DialTLSSkipVerify(rdbg.skipVerify),
				rgo.DialDatabase(rdbg.database),
			)
			return connDial, errDial
		},
		TestOnBorrow: func(c rgo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	return pool, pool.Get().Err()
}

func (rdbg *Redigo) Acquire(ctx context.Context) (conn rgo.Conn, err error) {
	if len(rdbg.addresses) > 1 {
		return rdbg.Cluster.Get(), nil
	} else {
		return rdbg.Pool.GetContext(ctx)
	}
}

func (rdbg *Redigo) prepareOptions(ctx context.Context, opt *RedigoPoolOptions) (retOpts RedigoPoolOptions) {
	if ctx != nil {
		rdbg.ctx = ctx
	}

	if len(opt.Addresses) > 0 {
		retOpts.Addresses = append(retOpts.Addresses, opt.Addresses...)
		rdbg.addresses = append(rdbg.addresses, opt.Addresses...)
	}

	if opt.Password != "" {
		retOpts.Password = opt.Password
		rdbg.password = opt.Password
	}

	if opt.ClientName != "" {
		retOpts.ClientName = opt.ClientName
		rdbg.clientName = opt.ClientName
	}

	if opt.Database > 0 {
		retOpts.Database = opt.Database
		rdbg.database = opt.Database
	} else {
		opt.Database = 0
		rdbg.database = 0
	}

	retOpts.TraceServiceName = "redis.db"
	if opt.TraceServiceName != "" {
		retOpts.TraceServiceName = opt.TraceServiceName
		rdbg.traceServiceName = opt.TraceServiceName
	}

	if opt.MaxIdle > 0 {
		retOpts.MaxIdle = opt.MaxIdle
		rdbg.maxIdle = opt.MaxIdle
	} else {
		retOpts.MaxIdle = MAX_IDLE
		rdbg.maxIdle = MAX_IDLE
	}

	if opt.MaxActive > 0 {
		retOpts.MaxActive = opt.MaxActive
		rdbg.maxActive = opt.MaxActive
	} else {
		retOpts.MaxActive = MAX_ACTIVE
		rdbg.maxActive = MAX_ACTIVE
	}

	if opt.UsageTLS {
		tlsConfig := &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: TRUE,
		}
		retOpts.UsageTLS = TRUE
		retOpts.TlsConfig = tlsConfig
		rdbg.usageTLS = TRUE
		rdbg.tlsConfig = tlsConfig
	} else {
		retOpts.UsageTLS = FALSE
		retOpts.SkipVerify = TRUE
		rdbg.usageTLS = FALSE
		rdbg.skipVerify = TRUE
	}

	retOpts.PoolSize = MAX_ACTIVE
	if opt.PoolSize > 0 {
		retOpts.PoolSize = opt.PoolSize
		rdbg.poolSize = opt.PoolSize
	}

	if opt.MaxConnLifetime > 0 {
		retOpts.MaxConnLifetime = opt.MaxConnLifetime
		rdbg.maxConnLifetime = opt.MaxConnLifetime
	} else {
		retOpts.MaxConnLifetime = time.Second * 3600
		rdbg.maxConnLifetime = time.Second * 3600
	}

	if opt.IdleTimeout > 0 {
		retOpts.IdleTimeout = opt.IdleTimeout
		rdbg.idleTimeout = opt.IdleTimeout
	} else {
		retOpts.IdleTimeout = time.Minute
		rdbg.idleTimeout = time.Minute
	}

	// retOpts.RouteByLatency = true
	return retOpts
}
