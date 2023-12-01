package main

import (
	"github.com/gofiber/fiber/v2"

	"github.com/fsvxavier/default-vertical-slice/pkg/httpserver/fiber/middleware"
)

func main() {
	app := fiber.New()

	// This here will appear as a label, one can also use
	// fiberprometheus.NewWith(servicename, namespace, subsystem )
	// or
	// NOTE: Following is not available in v1
	// labels := map[string]string{"custom_label1":"custom_value1", "custom_label2":"custom_value2"}
	// fiberprometheus.NewWithLabels(labels, namespace, subsystem )
	prometheus := middleware.NewPrometheus("my-service-name")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	app.Post("/some", func(c *fiber.Ctx) error {
		return c.SendString("Welcome!")
	})

	app.Listen(":8080")
}
