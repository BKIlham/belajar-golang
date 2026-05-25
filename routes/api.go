package routes

import (
	"cobago/controllers"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userController *controllers.UserController) {
	api := app.Group("/api")
	
	api.Post("/users", userController.Register)
	api.Get("/users", userController.GetAll)
	api.Get("/users/:id", userController.GetProfile)
	api.Patch("/users/:id", userController.Update)
	api.Delete("/users/:id", userController.SoftDelete)
	api.Delete("/users/:id/permanent", userController.HardDelete)
	api.Post("/users/:id/restore", userController.Restore)
}