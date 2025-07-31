package main

import (
	"Repos/ticketing-go/migrations"
	"Repos/ticketing-go/routes"
	"Repos/ticketing-go/storage"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	// Initialize route handlers
	merchantRoutes := &routes.MerchantRoutes{DB: r.DB}
	customerRoutes := &routes.CustomerRoutes{DB: r.DB}
	
	// Setup routes
	merchantRoutes.SetupRoutes(app)
	customerRoutes.SetupRoutes(app)
	// Add other route handlers here (TicketRoutes, etc.)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading environment variables:", err.Error())
	}

	config := &storage.Config{
		Host:    os.Getenv("DB_HOST"),
		Port:    os.Getenv("DB_PORT"),
		Name:    os.Getenv("DB_NAME"),
		User:    os.Getenv("DB_USER"),
		Pass:    os.Getenv("DB_PASS"),
		SSLMode: os.Getenv("SSL_MODE"),
	}
	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("Error connecting to database:", err.Error())
	}

	err = migrations.MigrateAll(db)
	if err != nil {
		log.Fatal("Error migrating database:", err.Error())
	}
	r := Repository{DB: db}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":3000")
}
