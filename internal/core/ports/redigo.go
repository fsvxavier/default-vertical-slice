package ports

import (
	"context"
)

// IRedigoRepository ...
type IRedigoRepository interface {
	SetRateCacheWitchoutTTL(ctx context.Context, key, value string) (err error)
	GetCache(ctx context.Context, key string) (cache string, err error)
	Get(ctx context.Context, key string) (string, error)
	HGet(ctx context.Context, hash, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Set(ctx context.Context, key, val string) error
	HSet(ctx context.Context, hash, key, val string) error
	Ping(ctx context.Context) error
	Reset() error
}
