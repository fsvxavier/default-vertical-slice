package redis

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	rgo "github.com/gomodule/redigo/redis"
)

type RedigoCache struct {
	ctx  context.Context
	conn rgo.Conn
}

func NewCache(options *RedigoPoolOptions) *RedigoCache {
	rgoInstance, _ := NewRedigo(options.Context, &RedigoPoolOptions{
		Addresses: options.Addresses,
		Password:  options.Password,
		Database:  options.Database,
	})

	conn, _ := rgoInstance.Acquire(options.Context)

	return &RedigoCache{
		ctx:  options.Context,
		conn: conn,
	}
}

func (r *RedigoCache) Close() error {
	return r.conn.Close()
}

func (r *RedigoCache) Get(key string) ([]byte, error) {
	val, err := rgo.String(r.conn.Do("GET", key, r.ctx))
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, err
	}
	return []byte(val), err
}

func (r *RedigoCache) Delete(key string) error {
	_, err := r.conn.Do("DEL", key, r.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *RedigoCache) Set(key string, val []byte, exp time.Duration) error {
	_, err := rgo.String(r.conn.Do("SET", key, val, r.ctx))
	if err != nil {
		return err
	}
	return nil
}

func (r *RedigoCache) Reset() error {
	return nil
}

// Interface conformance.
var _ fiber.Storage = (*RedigoCache)(nil)
