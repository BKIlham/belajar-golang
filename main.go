package main

import (
	"cobago/config"
	"cobago/controllers"
	"cobago/middlewares"
	"cobago/repositories"
	"cobago/routes"
	"cobago/services"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan Environment System")
	}

	db := config.ConnectDB()
	minioClient := config.ConnectMinIO()
	redisClient := config.ConnectRedis()

	userRepo := repositories.NewUserRepository(db)
	storageService := services.NewStorageService(minioClient)
	userService := services.NewUserService(userRepo, redisClient)
	userController := controllers.NewUserController(userService, storageService)
	
	authService := services.NewAuthService(userRepo, redisClient)
	authController := controllers.NewAuthController(authService)
	
	app := fiber.New()

	middlewares.SetupMiddlewares(app, redisClient)
	routes.SetupRoutes(app, userController, authController)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
