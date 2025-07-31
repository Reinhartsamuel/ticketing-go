package routes

import (
	"Repos/ticketing-go/models"
	"fmt"
	"reflect"

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
	app.Patch("/merchants/:id", m.UpdateMerchant)
	app.Delete("/merchants/:id", m.DeleteMerchant)
}

func (m *MerchantRoutes) CreateMerchant(c *fiber.Ctx) error {
	merchant := models.Merchant{}
	err := c.BodyParser(&merchant)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Dynamic validation
	requiredFields := map[string]string{
		"MerchantName":   "merchant name",
		"MerchantWallet": "merchant wallet",
		"MerchantType":   "merchant type",
	}

	for field, displayName := range requiredFields {
		val := reflect.ValueOf(merchant).FieldByName(field)
		if val.Kind() == reflect.String && val.String() == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("%s is required", displayName),
			})
		}
	}

	// Create merchant
	err = m.DB.Create(&merchant).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create merchant",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"message":  "Merchant created successfully",
		"merchant": merchant,
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
			"message": "Failed to delete merchant",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Merchant deleted successfully",
	})
}
