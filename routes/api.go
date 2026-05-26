package routes

import (
	"cobago/controllers"
	"cobago/middlewares"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, userController *controllers.UserController, authController *controllers.AuthController) {
	api := app.Group("/api")
	
	// 🔓 PUBLIC ENDPOINTS (Tanpa Login)
	auth := api.Group("/auth")
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.Refresh)

	api.Get("/users", userController.GetAll)
	api.Post("/users", userController.Register)
	api.Get("/users/:id", userController.GetProfile)

	// 🔒 PROTECTED ENDPOINTS (Wajib Menyertakan Access Token Valid)
	api.Patch("/users/:id", middlewares.Protected(), userController.Update)
	api.Delete("/users/:id", middlewares.Protected(), userController.SoftDelete)
	api.Delete("/users/:id/permanent", middlewares.Protected(), userController.HardDelete)
	api.Post("/users/:id/restore", middlewares.Protected(), userController.Restore)
	
	auth.Post("/logout", middlewares.Protected(), authController.Logout)
}