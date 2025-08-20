package main

import (
	"log"
	"os"

	"Repos/ticketing-go/migrations"
	"Repos/ticketing-go/routes"
	"Repos/ticketing-go/storage"
	// "Repos/ticketing-go/workers"

	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

// Repository struct to hold the database connection
type Repository struct {
	DB *gorm.DB
}

// SetupRoutes initializes all the route handlers and sets up the application routes.
func (r *Repository) SetupRoutes(app *fiber.App) {
	// Initialize route handlers
	merchantRoutes := &routes.MerchantRoutes{DB: r.DB}
	customerRoutes := &routes.CustomerRoutes{DB: r.DB}
	eventRoutes := &routes.EventRoutes{DB: r.DB}
	reservationRoutes := &routes.ReservationRoutes{DB: r.DB}

	// Setup routes
	merchantRoutes.SetupRoutes(app)
	customerRoutes.SetupRoutes(app)
	eventRoutes.SetupRoutes(app)
	reservationRoutes.SetupRoutes(app)
}

// main function to start the application
func main() {
	// ❌ Removed the line that loads the .env file.
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading environment variables:", err.Error())
	// }

	// Get environment variables directly from the system environment
	// Railway automatically injects these from the dashboard.
	config := &storage.Config{
		Host:    os.Getenv("DB_HOST"),
		Port:    os.Getenv("DB_PORT"),
		Name:    os.Getenv("DB_NAME"),
		User:    os.Getenv("DB_USER"),
		Pass:    os.Getenv("DB_PASS"),
		SSLMode: os.Getenv("SSL_MODE"),
	}

	// Check if any of the critical environment variables are missing
	if config.Host == "" || config.Port == "" || config.Name == "" || config.User == "" || config.Pass == "" {
		log.Fatal("Error: One or more database environment variables are not set. Check your Railway dashboard.")
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("Error connecting to database:", err.Error())
	}

	// ✅ Start poller in the background
	// go workers.StartVerificationPoller(db)

	err = migrations.MigrateAll(db)
	if err != nil {
		log.Fatal("Error migrating database:", err.Error())
	}
	r := Repository{DB: db}

	app := fiber.New()
	r.SetupRoutes(app)

	// Start the Fiber application on the port provided by Railway
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default port if not set
	}
	log.Fatal(app.Listen(":" + port))
}
