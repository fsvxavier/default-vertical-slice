package repositories

import (
	"context"

	rgo "github.com/gomodule/redigo/redis"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/fsvxavier/default-vertical-slice/internal/core/ports"
)

//go:generate ifacemaker -D -f redis.go -s RedisRepository -i IRedisRepository -p repositories -o ../interfaces/repositories/redis.go

type RedigoRepository struct {
	Conn rgo.Conn
}

const (
	RDB_RATE_CACHE = 0
)

func NewRedigoRepository(conn rgo.Conn) ports.IRedigoRepository {
	return &RedigoRepository{
		Conn: conn,
	}
}

func (rdb *RedigoRepository) SetRateCacheWitchoutTTL(ctx context.Context, key, value string) (err error) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.SetRateCacheWitchoutTTL")
	defer span.Finish()

	_, err = rgo.String(rdb.Conn.Do("SET", key, value, ctxs))
	if err != nil {
		return err
	}

	return nil
}

func (rdb *RedigoRepository) GetCache(ctx context.Context, key string) (cache string, err error) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.GetCache")
	defer span.Finish()

	cache, err = rgo.String(rdb.Conn.Do("GET", key, ctxs))
	if err != nil {
		if err.Error() == "redis: nil" {
			return "", nil
		}
		return "", err
	}
	return cache, nil
}

func (rdb *RedigoRepository) Get(ctx context.Context, key string) (string, error) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.Get")
	defer span.Finish()
	val, err := rgo.String(rdb.Conn.Do("GET", key, ctxs))
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rdb *RedigoRepository) HGet(ctx context.Context, hash, key string) (string, error) {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.HGet")
	defer span.Finish()

	val, err := rgo.String(rdb.Conn.Do("HGET", hash, key, ctxs))
	return val, err
}

func (rdb *RedigoRepository) Delete(ctx context.Context, key string) error {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.Get")
	defer span.Finish()

	_, err := rdb.Conn.Do("DEL", key, ctxs)
	if err != nil {
		return err
	}
	return nil
}

func (rdb *RedigoRepository) Set(ctx context.Context, key, val string) error {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.Set")
	defer span.Finish()

	_, err := rgo.String(rdb.Conn.Do("SET", key, val, ctxs))
	if err != nil {
		return err
	}
	return nil
}

func (rdb *RedigoRepository) HSet(ctx context.Context, hash, key, val string) error {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.HSet")
	defer span.Finish()

	_, err := rgo.String(rdb.Conn.Do("HSET", key, val, ctxs))
	if err != nil {
		return err
	}
	return nil
}

func (rdb *RedigoRepository) Ping(ctx context.Context) error {
	span, ctxs := tracer.StartSpanFromContext(ctx, "RedisRepository.Ping")
	defer span.Finish()

	_, err := rgo.String(rdb.Conn.Do("PING", ctxs))
	if err != nil {
		return err
	}
	return nil
}

func (rdb *RedigoRepository) Reset() error {
	return nil
}
