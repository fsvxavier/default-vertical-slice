package routering

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	handlers "github.com/fsvxavier/default-vertical-slice/internal/features/healthcheck/adapters/controllers/http"
	"github.com/fsvxavier/default-vertical-slice/pkg/database/redis"
	logger "github.com/fsvxavier/default-vertical-slice/pkg/logger/zap"
)

type Routes struct {
	App   *fiber.App
	Db    *pgxpool.Pool
	Redis *redis.Redigo
}

func NewRoutes(app *fiber.App, db *pgxpool.Pool, rdb *redis.Redigo) Routes {
	return Routes{
		App:   app,
		Db:    db,
		Redis: rdb,
	}
}

func (r Routes) SetupRoutes() {
	router := r.App.Group("/")

	// Health Routes
	r.healthRoutes(router)
}

func (r Routes) healthRoutes(router fiber.Router) {
	health := router.Group("/")

	hcHandlers := handlers.NewHealthCheckController(r.Db, r.Redis)

	health.Get("/health", func(ctx *fiber.Ctx) error {
		logger.Debug(ctx.UserContext(), ctx.Get("X-Kubernetes-Probe"))

		switch ctx.Get("X-Kubernetes-Probe") {
		case "ready":
			return hcHandlers.GetHealthcheck(ctx)
		case "live":
			return hcHandlers.GetHealthcheck(ctx)
		default:
			return hcHandlers.GetHealthcheck(ctx)
		}
	})
}
