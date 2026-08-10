package server

import (
	"github.com/Flussen/swagger-fiber-v3"
	"github.com/gofiber/fiber/v3"

	_ "github.com/ALLAN-star-glitch/nuruvent-backend/docs"
)

func SetupSwagger(app *fiber.App) {
	// Serve Swagger UI with default config
	app.Get("/swagger/*", swagger.HandlerDefault)

	// Redirect /docs to /swagger/index.html
	app.Get("/docs", func(c fiber.Ctx) error {
		return c.Redirect().To("/swagger/index.html")
	})

	// Redirect /swagger to /swagger/index.html
	app.Get("/swagger", func(c fiber.Ctx) error {
		return c.Redirect().To("/swagger/index.html")
	})
}