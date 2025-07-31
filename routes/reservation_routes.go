package routes

import (
	"Repos/ticketing-go/models"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ReservationRoutes struct {
	DB *gorm.DB
}

func (r *ReservationRoutes) SetupRoutes(app *fiber.App) {
	app.Post("/reservations", r.CreateReservation)
	app.Get("/reservations", r.GetReservations)
	app.Get("/reservations/:id", r.GetReservationByID)
	app.Patch("/reservations/:id", r.UpdateReservation)
}

func (r *ReservationRoutes) CreateReservation(c *fiber.Ctx) error {
	reservation := models.Reservation{}
	err := c.BodyParser(&reservation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	requiredFields := map[string]string{
		"UserID":  "user ID",
		"EventID": "event ID",
	}

	for field, displayName := range requiredFields {
		val := reflect.ValueOf(reservation).FieldByName(field)
		if (val.Kind() == reflect.String && val.String() == "") ||
			(val.Kind() == reflect.Uint && val.Uint() == 0) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("%s is required", displayName),
			})
		}
	}

	err = r.DB.Create(&reservation).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create reservation",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"message":     "Reservation created successfully",
		"reservation": reservation,
	})
}

func (r *ReservationRoutes) GetReservations(c *fiber.Ctx) error {
	var reservations []models.Reservation
	if err := r.DB.Find(&reservations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch reservations",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(reservations)
}

func (r *ReservationRoutes) GetReservationByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var reservation models.Reservation
	if err := r.DB.First(&reservation, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Reservation not found",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(reservation)
}

func (r *ReservationRoutes) UpdateReservation(c *fiber.Ctx) error {
	id := c.Params("id")
	var reservation models.Reservation
	if err := r.DB.First(&reservation, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Reservation not found",
			"error":   err.Error(),
		})
	}

	updateData := make(map[string]interface{})
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if err := r.DB.Model(&reservation).Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update reservation",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message":     "Reservation updated successfully",
		"reservation": reservation,
	})
}
