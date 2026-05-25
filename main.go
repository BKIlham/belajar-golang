package main

import (
	"cobago/config"
	"cobago/controllers"
	"cobago/middlewares"
	"cobago/repositories"
	"cobago/services"
	"cobago/routes"
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

	userRepo := repositories.NewUserRepository(db)
	storageService := services.NewStorageService(minioClient)
	userService := services.NewUserService(userRepo)
	userController := controllers.NewUserController(userService, storageService)

	app := fiber.New()

	middlewares.SetupMiddlewares(app) 
	routes.SetupRoutes(app, userController)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Fatal(app.Listen(":" + port))
}
