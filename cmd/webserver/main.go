package webserver

import (
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/fsvxavier/default-vertical-slice/cmd/webserver/routering"
	. "github.com/fsvxavier/default-vertical-slice/config"
	"github.com/fsvxavier/default-vertical-slice/pkg/database/gpgx"
	"github.com/fsvxavier/default-vertical-slice/pkg/database/redis"
	"github.com/fsvxavier/default-vertical-slice/pkg/httpserver/fiber"
	logger "github.com/fsvxavier/default-vertical-slice/pkg/logger/zap"
	"github.com/fsvxavier/default-vertical-slice/pkg/tracing/datadog"
)

const (
	AppName                      = "munin-exchange-rate-api"
	HTTP                         = "HTTP"
	TRUE                         = true
	FALSE                        = false
	DEFAULT_MAX_CONS             = "20"
	DEFAULT_MIN_CONS             = "1"
	DEFAULT_CONN_LIFE_TIME       = "3600"
	DEFAULT_CONN_IDLE_TIME       = "120"
	DEFAULT_RDB_MAX_ACTIVE_CONNS = "1000"
	DEFAULT_RDB_DATABASE         = "0"
	DEFAULT_RDB_MAX_IDLE_CONNS   = "300"
	DEFAULT_RDB_CLIENT_NAME      = "munin-exchange-rate-api"
)

func initDatabase(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "main.initDatabase")
	defer span.Finish()

	// Define max connections
	defaultMaxConns := DEFAULT_MAX_CONS
	if os.Getenv("DB_MAX_CONNS") != "" {
		defaultMaxConns = os.Getenv("DB_MAX_CONNS")
	}
	maxConns, err := strconv.ParseInt(defaultMaxConns, 10, 64)
	if err != nil {
		logger.Error(ctx, err.Error())
	}

	// Define min connections
	defaultMinConns := DEFAULT_MIN_CONS
	if os.Getenv("DB_MIN_CONNS") != "" {
		defaultMinConns = os.Getenv("DB_MIN_CONNS")
	}
	minConns, err := strconv.ParseInt(defaultMinConns, 10, 64)
	if err != nil {
		logger.Error(ctx, err.Error())
	}

	// Define life time connections
	defaultLifeConns := DEFAULT_CONN_LIFE_TIME
	if os.Getenv("DB_LIFE_TIME_CONNS") != "" {
		defaultLifeConns = os.Getenv("DB_LIFE_TIME_CONNS")
	}
	lifeTimeConns, err := time.ParseDuration(defaultLifeConns + "s")
	if err != nil {
		logger.Error(ctx, err.Error())
	}

	// Define idle life time connections
	defaultIdleConns := DEFAULT_CONN_IDLE_TIME
	if os.Getenv("DB_IDLE_TIME_CONNS") != "" {
		defaultIdleConns = os.Getenv("DB_IDLE_TIME_CONNS")
	}
	idleTimeConns, err := time.ParseDuration(defaultIdleConns + "s")
	if err != nil {
		logger.Error(ctx, err.Error())
	}

	multiTenantRep := strings.ToUpper(cfg.Application.Drivers) == HTTP

	pool := gpgx.NewPgConnection().
		SetMaxConns(int32(maxConns)).
		SetMinConns(int32(minConns)).
		SetMaxConnLifetime(lifeTimeConns).
		SetMaxConnIdleTime(idleTimeConns).
		SetDatadogEnable(cfg.Datadog.Enabled).
		SetQueryTracerEnabled(cfg.Database.QueryTracer).
		SetMultiTenantEnabled(cfg.Database.MultiTenant).
		SetMultiTenantRepEnabled(multiTenantRep)

	err = pool.NewPool(ctxs, cfg.Database.Connection.Url)
	if err != nil {
		logger.Fatal(ctxs, "Error to create a new pool database - "+err.Error())
	}

	if cfg.Database.Ping {
		err = pool.Pool().Ping(ctxs)
		if err != nil {
			logger.Fatal(ctxs, "Error to ping database - "+err.Error())
		}
	}

	return pool.Pool(), nil
}

func initLogger(ctx context.Context, outout io.Writer, cfg *Config) (loging *logger.Logger) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "main.initLogger")
	defer span.Finish()

	return logger.NewLogger().WithLevel(cfg.Log.Level).WithContext(ctxs).SetOutput(outout)
}

func initRedis(ctx context.Context, cfg *Config) (rdb redis.Redigo, err error) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "main.initRedis")
	defer span.Finish()

	// Define max active connections
	defaultMaxActiveConns := DEFAULT_RDB_MAX_ACTIVE_CONNS
	if os.Getenv("RDB_MAX_ACTIVE_CONNS") != "" {
		defaultMaxActiveConns = os.Getenv("RDB_MAX_ACTIVE_CONNS")
	}
	maxActive, err := strconv.ParseInt(defaultMaxActiveConns, 10, 64)
	if err != nil {
		logger.Fatal(ctxs, "Error to ParseInt maxActive - "+err.Error())
		return redis.Redigo{}, err
	}

	// Define max active connections
	defaultDatabase := DEFAULT_RDB_DATABASE
	if os.Getenv("RDB_DATABASE") != "" {
		defaultDatabase = os.Getenv("RDB_DATABASE")
	}
	rdbDatabase, err := strconv.ParseInt(defaultDatabase, 10, 64)
	if err != nil {
		logger.Fatal(ctxs, "Error to ParseInt rdbDatabase - "+err.Error())
		return redis.Redigo{}, err
	}

	// Define max active connections
	defaultMaxIdleConns := DEFAULT_RDB_MAX_IDLE_CONNS
	if os.Getenv("RDB_MAX_IDLE_CONNS") != "" {
		defaultMaxIdleConns = os.Getenv("RDB_MAX_IDLE_CONNS")
	}
	maxIdleConns, err := strconv.ParseInt(defaultMaxIdleConns, 10, 64)
	if err != nil {
		logger.Fatal(ctxs, "Error to ParseInt maxIdleConns - "+err.Error())
		return redis.Redigo{}, err
	}

	opt := &redis.RedigoPoolOptions{
		Addresses:        strings.Split(cfg.Redis.Addresses, ","),
		MaxIdle:          int(maxIdleConns),
		MaxActive:        int(maxActive),
		MaxConnLifetime:  time.Second * 3600,
		Database:         int(rdbDatabase),
		ClientName:       cfg.Redis.ClientName,
		Password:         cfg.Redis.Password,
		UsageTLS:         cfg.Redis.UsageTLS,
		TraceServiceName: cfg.Redis.TraceServiceName,
	}

	rdb, err = redis.NewRedigo(ctxs, opt)
	if err != nil {
		return rdb, err
	}

	return rdb, nil
}

func Run() {
	ctxs := context.TODO()

	cfg := NewConfig()

	if cfg.Datadog.Enabled {
		datadog.StartTracing()
		defer datadog.StopTracing()

		// Start a root span.
		var span ddtrace.Span
		span, ctxs = tracer.StartSpanFromContext(context.Background(), "main")
		defer span.Finish()

		if cfg.Datadog.Profile {
			err := datadog.StartProfiling()
			if err != nil {
				logger.Error(ctxs, err.Error())
			}
			defer datadog.StopProfiling()
		}
	}

	dbPool, err := initDatabase(ctxs, cfg)
	if err != nil {
		logger.Panic(ctxs, "Error to connect Database - "+err.Error())
	}
	defer dbPool.Close()

	rdb, err := initRedis(ctxs, cfg)
	if err != nil {
		logger.Panic(ctxs, "Error to connect Redis - "+err.Error())
	}
	defer rdb.Pool.Close()

	httpServer := fiber.FiberEngine{}

	httpServer.NewWebserver(cfg.Http.Port)
	router := routering.NewRoutes(httpServer.GetApp(), dbPool, &rdb)
	router.SetupRoutes()
	httpServer.Router(router.App)
	httpServer.Run()
}
