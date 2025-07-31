package main

import (
	"Repos/ticketing-go/migrations"
	"Repos/ticketing-go/models"
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
	app.Post("/merchants", r.CreateMerchant)
	app.Get("/merchants", r.GetMerchants)
	app.Get("/merchants/:id", r.GetMerchantByID)
	app.Put("/merchants/:id", r.UpdateMerchant)
}

func (r *Repository) CreateMerchant(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := c.BodyParser(&merchant)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	err = r.DB.Create(&merchant).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create merchant",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Merchant created successfully",
	})
}

func (r *Repository) GetMerchants(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := r.DB.First(&merchant).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Merchant not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"merchant": merchant,
	})
}

func (r *Repository) GetMerchantByID(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := r.DB.First(&merchant, c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Merchant not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"merchant": merchant,
	})
}

func (r *Repository) UpdateMerchant(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := c.BodyParser(&merchant)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	err = r.DB.Save(&merchant).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update merchant",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Merchant updated successfully",
	})
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
