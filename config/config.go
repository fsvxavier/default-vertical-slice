package config

import (
	"os"

	cstt "github.com/fsvxavier/default-vertical-slice/internal/core/commons/constants"
)

type Config struct {
	Aws         `json:"aws,omitempty"`
	Database    `json:"database,omitempty"`
	Datadog     `json:"datadog,omitempty"`
	Application `json:"application,omitempty"`
	Log         `json:"log,omitempty"`
	Http        `json:"http,omitempty"`
	Redis       `json:"redis,omitempty"`
}

type Application struct {
	Env            string `env:"ENV"             json:"env,omitempty"`
	Name           string `env:"APP_NAME"        json:"app_name,omitempty"`
	Path           string `env:"HOME"            json:"path,omitempty"`
	AppVersion     string `env:"APP_VERSION"     json:"app_version,omitempty"`
	Drivers        string `env:"EVENT_DRIVERS"   json:"drivers,omitempty"`
	Port           string `env:"PORT"            json:"port,omitempty"`
	GitCredentials string `env:"GIT_CREDENTIALS" json:"git_credentials,omitempty"`
	Pprof          bool   `env:"PPROF_ENABLED"   json:"pprof,omitempty"`
}

type Http struct {
	Port            string `env:"HTTP_PORT"              json:"http_port,omitempty"`
	Host            string `env:"HTTP_HOST"              json:"http_host,omitempty"`
	Network         string `env:"HTTP_NETWORK"           json:"http_network,omitempty"`
	Concurrency     string `env:"HTTP_CONCURRENCY"       json:"http_concurrency,omitempty"`
	Metrics         bool   `env:"HTTP_METRICS_ENABLED"   envDefault:"false"                json:"http_metrics_enabled,omitempty"`
	Prefork         bool   `env:"HTTP_PREFORK"           envDefault:"false"                json:"http_prefork,omitempty"`
	Rmu             bool   `env:"HTTP_RMU"               envDefault:"true"                 json:"http_rmu,omitempty"`
	DisableStartMsg bool   `env:"HTTP_DISABLE_START_MSG" envDefault:"true"                 json:"http_disable_start_msg,omitempty"`
}

type Datadog struct {
	Service       string `env:"DD_SERVICE"          envDefault:"munin-exchange-rate-api"    json:"dd_service,omitempty"`
	ServiceDb     string `env:"DD_SERVICE_DB"       envDefault:"munin-exchange-rate-api.db" json:"dd_service_db,omitempty"`
	Env           string `env:"DD_ENV"              envDefault:"hml"                        json:"dd_env,omitempty"`
	Version       string `env:"DD_VERSION"          envDefault:"latest"                     json:"dd_version,omitempty"`
	AgentHost     string `env:"DD_AGENT_HOST"       json:"dd_agent_host,omitempty"`
	AgentPort     string `env:"DD_TRACE_AGENT_PORT" json:"dd_agent_port,omitempty"`
	Enabled       bool   `env:"DD_ENABLED"          envDefault:"true"                       json:"dd_enabled,omitempty"`
	Profile       bool   `env:"DD_PROFILE"          envDefault:"false"                      json:"dd_profile,omitempty"`
	ApmEnable     bool   `env:"DD_APM_ENABLED"      json:"dd_apm_enable,omitempty"`
	LogsInjection bool   `env:"DD_LOGS_INJECTION"   json:"dd_logs_injection,omitempty"`
	TraceEnable   bool   `env:"DD_TRACE_ENABLED"    json:"dd_trace_enable,omitempty"`
}

type Log struct {
	Level   string `env:"LOG_LEVEL"      json:"log_level,omitempty"`
	Enabled bool   `env:"DD_APM_ENABLED" json:"log_enabled,omitempty"`
}

type Database struct {
	Connection  Connection `json:"connection"`
	Driver      string     `env:"DB_DIVER"               json:"db_driver,omitempty"`
	Type        string     `env:"DB_TYPE"                json:"db_type,omitempty"`
	MaxConns    string     `env:"DB_MAX_CONNS"           json:"db_max_conns,omitempty"`
	QueryTracer bool       `env:"DB_QUERY_TRACER"        json:"db_query_tracer,omitempty"`
	MultiTenant bool       `env:"DB_MULTI_TENANT_ENABLE" json:"db_multi_tenant,omitempty"`
	Ping        bool       `env:"DB_EXECUTE_PING"        json:"db_execute_ping,omitempty"`
}

type Redis struct {
	MaxIdleConns     string `env:"RDB_MAX_IDLE_CONNS"   json:"rdb_max_idle_conns,omitempty"`
	ClientName       string `env:"RDB_CLIENT_NAME"      json:"rdb_client_name,omitempty"`
	Username         string `env:"RDB_USERNAME"         json:"rdb_username,omitempty"`
	Password         string `env:"RDB_PASSWORD"         json:"rdb_password,omitempty"`
	MaxRetries       string `env:"RDB_MAX_RETRIES"      json:"rdb_max_retries,omitempty"`
	MinIdleConns     string `env:"RDB_MIN_IDLE_CONNS"   json:"rdb_min_idle_conns,omitempty"`
	Addresses        string `env:"RDB_ADDRESSES"        json:"rdb_addresses,omitempty"`
	MaxActiveConns   string `env:"RDB_MAX_ACTIVE_CONNS" json:"rdb_max_active_conns,omitempty"`
	PoolSize         string `env:"RDB_POOL_SIZE"        json:"rdb_pool_size,omitempty"`
	DatabaseDefault  string `env:"RDB_DATABASE_DEFAULT" json:"rdb_database_default,omitempty"`
	TraceServiceName string `env:"RDB_DD_SERVICE_DB"    json:"rdb_dd_service_db,omitempty"`
	Ping             bool   `env:"RDB_EXECUTE_PING"     json:"rdb_execute_ping,omitempty"`
	UsageTLS         bool   `env:"RDB_USAGE_TLS"        json:"rdb_usage_tls,omitempty"`
}

type Connection struct {
	Url      string `env:"DB_URL"      json:"db_url,omitempty"`
	Host     string `env:"DB_HOST"     json:"db_host,omitempty"`
	Port     string `env:"DB_PORT"     json:"db_port,omitempty"`
	Username string `env:"DB_USERNAME" json:"db_username,omitempty"`
	Password string `env:"DB_PASSWORD" json:"db_password,omitempty"`
	DbName   string `env:"DB_NAME"     json:"db_name,omitempty"`
	Schema   string `env:"DB_SCHEMA"   json:"db_schema,omitempty"`
}

type Aws struct {
	Profile string `env:"AWS_PROFILE" json:"aws_profile"`
	Region  string `env:"AWS_REGION"  json:"aws_region"`
}

func NewConfig() *Config {
	application := cfgApplication()
	log := cfgLog()
	database := cfgDatabase()
	http := cfgHttp()
	aws := cfgAws()
	datadog := cfgDatadog()
	redis := cfgRedis()

	return &Config{
		Application: *application,
		Log:         *log,
		Database:    *database,
		Http:        *http,
		Aws:         *aws,
		Datadog:     *datadog,
		Redis:       *redis,
	}
}

func cfgApplication() *Application {
	return &Application{
		Env:            os.Getenv("ENV"),
		Name:           os.Getenv("APP_NAME"),
		Path:           os.Getenv("HOME"),
		AppVersion:     os.Getenv("APP_VERSION"),
		Drivers:        os.Getenv("EVENT_DRIVERS"),
		Port:           os.Getenv("PORT"),
		GitCredentials: os.Getenv("GIT_CREDENTIALS"),
		Pprof:          os.Getenv("PPROF_ENABLED") == cstt.STR_TRUE,
	}
}

func cfgLog() *Log {
	return &Log{
		Level:   os.Getenv("LOG_LEVEL"),
		Enabled: os.Getenv("DD_APM_ENABLED") == cstt.STR_TRUE,
	}
}

func cfgDatabase() *Database {
	cfgDatabase := Database{
		Driver:      os.Getenv("DB_DIVER"),
		Type:        os.Getenv("DB_TYPE"),
		MaxConns:    os.Getenv("DB_MAX_CONNS"),
		QueryTracer: os.Getenv("DB_QUERY_TRACER") == cstt.STR_TRUE,
		MultiTenant: os.Getenv("DB_MULTI_TENANT_ENABLE") == cstt.STR_TRUE,
		Ping:        os.Getenv("DB_EXECUTE_PING") == cstt.STR_TRUE,
	}

	// Database config from separate parameters
	dbh := os.Getenv("DB_HOST")
	if len(dbh) > 0 {
		cfgDatabase.Connection = Connection{
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			Username: os.Getenv("DB_USERNAME"),
			Password: os.Getenv("DB_PASSWORD"),
			DbName:   os.Getenv("DB_NAME"),
			Schema:   os.Getenv("DB_SCHEMA"),
		}

		str := "postgres://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable&search_path=${DB_SCHEMA}"
		cfgDatabase.Connection.Url = os.ExpandEnv(str)
	}

	return &cfgDatabase
}

func cfgHttp() *Http {
	return &Http{
		Port:            os.Getenv("HTTP_PORT"),
		Host:            os.Getenv("HTTP_HOST"),
		Network:         os.Getenv("HTTP_NETWORK"),
		Concurrency:     os.Getenv("HTTP_CONCURRENCY"),
		Metrics:         os.Getenv("HTTP_METRICS_ENABLED") == cstt.STR_TRUE,
		Prefork:         os.Getenv("HTTP_PREFORK") == cstt.STR_TRUE,
		Rmu:             os.Getenv("HTTP_RMU") == cstt.STR_TRUE,
		DisableStartMsg: os.Getenv("HTTP_DISABLE_START_MSG") == cstt.STR_TRUE,
	}
}

func cfgAws() *Aws {
	return &Aws{
		Region:  os.Getenv("AWS_REGION"),
		Profile: os.Getenv("AWS_PROFILE"),
	}
}

func cfgDatadog() *Datadog {
	return &Datadog{
		Service:       os.Getenv("DD_SERVICE"),
		ServiceDb:     os.Getenv("DD_SERVICE_DB"),
		Env:           os.Getenv("DD_ENV"),
		Version:       os.Getenv("DD_VERSION"),
		AgentHost:     os.Getenv("DD_AGENT_HOST"),
		AgentPort:     os.Getenv("DD_TRACE_AGENT_PORT"),
		Enabled:       os.Getenv("DATADOG_ENABLED") == cstt.STR_TRUE,
		Profile:       os.Getenv("DATADOG_PROFILE") == cstt.STR_TRUE,
		ApmEnable:     os.Getenv("DD_APM_ENABLED") == cstt.STR_TRUE,
		LogsInjection: os.Getenv("DD_LOGS_INJECTION") == cstt.STR_TRUE,
		TraceEnable:   os.Getenv("DD_TRACE_ENABLED") == cstt.STR_TRUE,
	}
}

func cfgRedis() *Redis {
	return &Redis{
		Addresses:        os.Getenv("RDB_ADDRESSES"),
		ClientName:       os.Getenv("RDB_CLIENT_NAME"),
		Username:         os.Getenv("RDB_USERNAME"),
		Password:         os.Getenv("RDB_PASSWORD"),
		MaxRetries:       os.Getenv("RDB_MAX_RETRIES"),
		MinIdleConns:     os.Getenv("RDB_MIN_IDLE_CONNS"),
		MaxIdleConns:     os.Getenv("RDB_MAX_IDLE_CONNS"),
		MaxActiveConns:   os.Getenv("RDB_MAX_ACTIVE_CONNS"),
		DatabaseDefault:  os.Getenv("RDB_DATABASE_DEFAULT"),
		PoolSize:         os.Getenv("RDB_POOL_SIZE"),
		Ping:             os.Getenv("RDB_EXECUTE_PING") == cstt.STR_TRUE,
		UsageTLS:         os.Getenv("RDB_USAGE_TLS") == cstt.STR_TRUE,
		TraceServiceName: os.Getenv("RDB_DD_SERVICE_DB"),
	}
}
