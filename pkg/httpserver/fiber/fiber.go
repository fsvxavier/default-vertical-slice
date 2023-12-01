package fiber

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/skip"
	"github.com/gofiber/swagger"
	fibertrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gofiber/fiber.v2"

	"github.com/fsvxavier/default-vertical-slice/pkg/httpserver/fiber/middleware"
	log "github.com/fsvxavier/default-vertical-slice/pkg/logger/zap"
)

type FiberEngine struct {
	app  *fiber.App
	port string
}

var healthcheckPath = func(c *fiber.Ctx) bool { return c.Path() == "/health" }

func (engine *FiberEngine) NewWebserver(serverPort string) {
	api := fiber.New(fiber.Config{
		ErrorHandler: middleware.ApplicationErrorHandler,
		// ReadBufferSize:        40960,
		DisableStartupMessage: os.Getenv("HTTP_DISABLE_START_MSG") == "true",
		Prefork:               os.Getenv("HTTP_PREFORK") == "true",
	})

	if os.Getenv("PPROF_ENABLED") == "true" {
		api.Use(pprof.New())
		api.Get("/metrics", monitor.New())
	}

	api.Use(skip.New(fibertrace.Middleware(), healthcheckPath))

	api.Use(recover.New(recover.Config{
		EnableStackTrace: os.Getenv("SHOW_STACK_TRACE") == "true",
	}))

	api.Use(skip.New(middleware.LoggerMiddleware(os.Stdout), healthcheckPath))
	api.Use(skip.New(middleware.TraceIdMiddleware, healthcheckPath))
	api.Use(skip.New(middleware.TenantIdMiddleware, healthcheckPath))
	api.Use(middleware.ContentTypeMiddleware("POST", fiber.MIMEApplicationJSON))

	if os.Getenv("HTTP_RATE_LIMIT_ENABLE") == "true" {
		api.Use(limiter.New(middleware.DefaultRateLimiterConfig))
	}

	engine.app = api
	engine.port = serverPort
}

func (engine *FiberEngine) GetApp() *fiber.App {
	return engine.app
}

func (engine *FiberEngine) GetPort() string {
	return engine.port
}

func (engine *FiberEngine) Run() {
	closed := make(chan bool, 1)

	log.Debugln(fmt.Sprintf("Listening on port %s", engine.port))

	engine.app.Listen(("0.0.0.0:" + engine.port))
	<-closed
}

func (engine *FiberEngine) Router(app *fiber.App) {
	app.Route("/docs/*", func(r fiber.Router) {
		r.Get("", swagger.New(swagger.Config{
			DocExpansion: "none",
		}))
	})

	app.All("/*", func(ctx *fiber.Ctx) error {
		ctx.Status(http.StatusForbidden)
		return ctx.JSON(fiber.Map{"message": "Forbidden"})
	})

	engine.app = app
}
