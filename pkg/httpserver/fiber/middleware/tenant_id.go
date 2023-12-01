package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

func TenantIdMiddleware(ctx *fiber.Ctx) error {
	tenantID := ctx.Get("Client-Id")

	c := ctx.UserContext()
	//lint:ignore SA1029 ignore this!
	c = context.WithValue(c, "tenant_id", tenantID)
	ctx.SetUserContext(c)

	return ctx.Next()
}
