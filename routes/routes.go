package routes

import (
	"github.com/PlenaFinance/solanacapi/controller"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {
	app.Use(logger.New())
	app.Post("/test", controller.Test)
}
