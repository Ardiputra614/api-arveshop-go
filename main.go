package main

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"api-arveshop-go/routes"
	"api-arveshop-go/utils"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	config.ConnectDB()
	config.DB.AutoMigrate(
		&models.User{},
		&models.Whatsapp{},
		&models.Transaction{},
		&models.Service{},
		&models.Product{},
		&models.ProductPasca{},
		&models.PaymentMethod{},
		&models.Category{},
		&models.ApplicationSetting{},		
	)

	 // Initialize Cloudinary - PENTING!
    if err := utils.InitCloudinary(); err != nil {
        log.Fatal("Failed to initialize Cloudinary: ", err)
    }

	routes.SetupRoutes(r)

	r.Run(":8080")
}
