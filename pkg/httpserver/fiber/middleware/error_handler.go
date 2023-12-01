package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fsvxavier/default-vertical-slice/pkg/httpserver/fiber/adapters"
	log "github.com/fsvxavier/default-vertical-slice/pkg/logger/zap"
)

func ApplicationErrorHandler(ctx *fiber.Ctx, err error) error {
	log.Errorln(err)

	return adapters.ResponseAdapter(
		ctx.UserContext(),
		ctx,
		adapters.ControllerResponse{Error: err},
	)
}
