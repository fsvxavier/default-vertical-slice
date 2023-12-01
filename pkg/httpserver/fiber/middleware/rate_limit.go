package middleware

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/fsvxavier/default-vertical-slice/internal/core/utils/helpers"
	"github.com/fsvxavier/default-vertical-slice/pkg/apierrors"
	"github.com/fsvxavier/default-vertical-slice/pkg/database/redis"
)

func rateLimiterStorage(ctx context.Context) fiber.Storage {
	redisHost := os.Getenv("REDIS_HOST")
	redisUsername := os.Getenv("REDIS_USERNAME")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	if redisHost != "" && redisUsername != "" && redisPassword != "" {
		addrs := strings.Split(redisHost, ",")
		return redis.NewCache(&redis.RedigoPoolOptions{
			Context:   ctx,
			Database:  1,
			Addresses: addrs,
			Password:  redisPassword,
		})
	}

	return nil
}

func rateLimitRPM() int {
	var requestsPerMinute int = 100

	env := os.Getenv("RATE_LIMIT_REQUESTS_PER_MINUTE")
	if value, err := strconv.Atoi(env); err == nil {
		requestsPerMinute = value
	}

	return requestsPerMinute
}

func rateLimitLockDuration() time.Duration {
	var lockDuration time.Duration = 1 * time.Minute

	env := os.Getenv("RATE_LIMIT_LOCK_DURATION_IN_MINUTES")
	if value, err := strconv.Atoi(env); err == nil {
		lockDuration = time.Duration(value) * time.Minute
	}

	return lockDuration
}

func rateLimiterConfig(storage fiber.Storage, requestsPerMinute int, lockDuration time.Duration) limiter.Config {
	return limiter.Config{
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		Storage:                storage,
		Max:                    requestsPerMinute,
		Expiration:             lockDuration,
		Next:                   func(c *fiber.Ctx) bool { return c.Path() == "/health" },
		KeyGenerator: func(c *fiber.Ctx) string {
			key := c.Get("Client-Id")

			if key == "" {
				return c.IP()
			}

			return key
		},
		LimitReached: func(c *fiber.Ctx) error {
			span, ddCtx := tracer.StartSpanFromContext(c.UserContext(), helpers.GetCurrentFuncName(), tracer.SpanType(ext.AppTypeCache))
			defer span.Finish()

			c.SetUserContext(ddCtx)

			c.Status(http.StatusTooManyRequests)
			err := apierrors.NewDockApiError(
				http.StatusUnsupportedMediaType,
				"429",
				"Too Many Requests",
			)

			err.SetId(c.Get("Trace-Id"))

			return c.JSON(err)
		},
	}
}

var DefaultRateLimiterConfig = rateLimiterConfig(
	rateLimiterStorage(context.TODO()),
	rateLimitRPM(),
	rateLimitLockDuration(),
)
