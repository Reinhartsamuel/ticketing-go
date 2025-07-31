package routes

import (
	"Repos/ticketing-go/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type MerchantRoutes struct {
	DB *gorm.DB
}

func (m *MerchantRoutes) SetupRoutes(app *fiber.App) {
	app.Post("/merchants", m.CreateMerchant)
	app.Get("/merchants", m.GetMerchants)
	app.Get("/merchants/:id", m.GetMerchantByID)
	app.Put("/merchants/:id", m.UpdateMerchant)
}

func (m *MerchantRoutes) CreateMerchant(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := c.BodyParser(&merchant)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	err = m.DB.Create(&merchant).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create merchant",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Merchant created successfully",
	})
}

func (m *MerchantRoutes) GetMerchants(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := m.DB.First(&merchant).Error
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

func (m *MerchantRoutes) GetMerchantByID(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := m.DB.First(&merchant, c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Merchant not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"merchant": merchant,
	})
}

func (m *MerchantRoutes) UpdateMerchant(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := c.BodyParser(&merchant)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	err = m.DB.Save(&merchant).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update merchant",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Merchant updated successfully",
	})
}
func (m *MerchantRoutes) DeleteMerchant(c *fiber.Ctx) error {
	err := m.DB.Delete(&models.Merchant{}, c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete customer",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Customer deleted successfully",
	})
}
