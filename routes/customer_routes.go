package routes

import (
	"Repos/ticketing-go/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CustomerRoutes struct {
	DB *gorm.DB
}

func (c *CustomerRoutes) SetupRoutes(app *fiber.App) {
	app.Post("/customers", c.CreateCustomer)
	app.Get("/customers", c.GetCustomers)
	app.Get("/customers/:id", c.GetCustomerByID)
	app.Patch("/customers/:id", c.UpdateCustomer)
	app.Delete("/customers/:id", c.DeleteCustomer)
}

func (cr *CustomerRoutes) CreateCustomer(c *fiber.Ctx) error {
	customer := models.Customer{}

	// Parse and validate request body
	if err := c.BodyParser(&customer); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Create customer record
	if err := cr.DB.Create(&customer).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"message":  "Customer created successfully",
		"customer": customer,
	})
}

func (cr *CustomerRoutes) GetCustomers(c *fiber.Ctx) error {
	var customers []models.Customer
	err := cr.DB.Find(&customers).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   err.Error(),
			"message": "Customers not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"customers": customers,
	})
}

func (cr *CustomerRoutes) GetCustomerByID(c *fiber.Ctx) error {
	customer := models.Customer{}
	err := cr.DB.First(&customer, "id = ?", c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Customer not found",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"customer": customer,
	})
}

func (cr *CustomerRoutes) UpdateCustomer(c *fiber.Ctx) error {
	customer := models.Customer{}
	err := c.BodyParser(&customer)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}
	err = cr.DB.Save(&customer).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update customer",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Customer updated successfully",
	})
}

func (cr *CustomerRoutes) DeleteCustomer(c *fiber.Ctx) error {
	err := cr.DB.Delete(&models.Customer{}, c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete customer",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Customer deleted successfully",
	})
}
