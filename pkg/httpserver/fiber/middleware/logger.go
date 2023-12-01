package middleware

import (
	"io"
	"time"

	"github.com/gofiber/fiber/v2"

	log "github.com/fsvxavier/default-vertical-slice/pkg/logger/zap"
)

func LoggerMiddleware(w io.Writer) func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		reqTime := time.Now()

		// Capture any error returned by the handler
		err := c.Next()

		log.Debugln(w, "Duration:", time.Since(reqTime), c.Protocol(), c.Method(), c.IP(), c.Path(), "-", c.Response().StatusCode())

		// logMessage := map[string]any{
		// 	"headers": map[string]string{
		// 		"trace-id":  c.Get("Trace-Id"),
		// 		"client-id": c.Get("Client-Id"),
		// 	},
		// 	"body": string(c.Body()),
		// }
		// fields := append(
		// 	[]log.Field{},
		// 	log.Reflect("data", logMessage),
		// )
		// log.Debug(c.UserContext(), "middleware log", fields...)

		return err
	}
}
