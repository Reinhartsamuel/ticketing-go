package routes

import (
	"Repos/ticketing-go/models"
	"fmt"
	"reflect"

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
	app.Put("/customers/:id", c.UpdateCustomer)
	app.Delete("/customers/:id", c.DeleteCustomer)
}

func (cr *CustomerRoutes) CreateCustomer(c *fiber.Ctx) error {
	customer := models.Customer{}
	err := c.BodyParser(&customer)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Dynamic validation
	requiredFields := map[string]string{
		"CustomerName":   "customer name",
		"CustomerWallet": "customer wallet",
		"CustomerType":   "customer type",
	}

	for field, displayName := range requiredFields {
		val := reflect.ValueOf(customer).FieldByName(field)
		if val.Kind() == reflect.String && val.String() == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("%s is required", displayName),
			})
		}
	}
	err = cr.DB.Create(&customer).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create customer",
		})
	}
	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Customer created successfully",
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
	err := cr.DB.First(&customer, c.Params("id")).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Customer not found",
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
