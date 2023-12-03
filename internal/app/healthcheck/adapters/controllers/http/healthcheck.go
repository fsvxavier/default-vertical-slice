package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/fsvxavier/default-vertical-slice/internal/app/healthcheck/core/services"
	"github.com/fsvxavier/default-vertical-slice/pkg/database/redis"
)

type healthcheckController struct {
	Db      *pgxpool.Pool
	RdbConn *redis.Redigo
}

func NewHealthCheckController(db *pgxpool.Pool, rdbConn *redis.Redigo) *healthcheckController {
	return &healthcheckController{
		Db:      db,
		RdbConn: rdbConn,
	}
}

// @Summary HealthCheck
// @Description HealthCheck API
// @Success 200
// @Router /healthcheck [get].
func (hcc *healthcheckController) GetHealthcheck(ctx *fiber.Ctx) (err error) {
	dbConn, err := hcc.Db.Acquire(context.TODO())
	defer dbConn.Release()
	if err != nil {
		return err
	}

	rdbConn, err := hcc.RdbConn.Acquire(context.TODO())
	defer rdbConn.Close()
	if err != nil {
		return err
	}

	hcService := services.NewHealthCheckService(dbConn, rdbConn)
	hcReturn, err := hcService.GetHealthcheck()
	if err != nil {
		ctx.SendStatus(500)
		return ctx.JSON(hcReturn)
	}

	ctx.SendStatus(200)
	return ctx.JSON(hcReturn)
}

func GetHealth(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{"status": "ok"})
}
