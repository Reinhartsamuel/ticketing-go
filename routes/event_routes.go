package routes

import (
	"Repos/ticketing-go/models"
	"fmt"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type EventRoutes struct {
	DB *gorm.DB
}

func (e *EventRoutes) SetupRoutes(app *fiber.App) {
	app.Post("/events", e.CreateEvent)
	app.Get("/events", e.GetEvents)
	app.Get("/events/:id", e.GetEventByID)
	app.Put("/events/:id", e.UpdateEvent)
}

func (e *EventRoutes) CreateEvent(c *fiber.Ctx) error {
	event := models.Event{}
	err := c.BodyParser(&event)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	requiredFields := map[string]string{
		"EventName":  "event name",
		"EventDate":  "event date",
		"EventTime":  "event time",
		"EventVenue": "event venue",
		"MerchantID": "merchant ID",
	}

	for field, displayName := range requiredFields {
		val := reflect.ValueOf(event).FieldByName(field)
		if (val.Kind() == reflect.String && val.String() == "") ||
			(val.Kind() == reflect.Struct && val.IsZero()) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("%s is required", displayName),
			})
		}
	}

	err = e.DB.Create(&event).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create event",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"message": "Event created successfully",
		"event":   event,
	})
}

func (e *EventRoutes) GetEvents(c *fiber.Ctx) error {
	var events []models.Event
	if err := e.DB.Find(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch events",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(events)
}

func (e *EventRoutes) GetEventByID(c *fiber.Ctx) error {
	id := c.Params("id")
	var event models.Event
	if err := e.DB.First(&event, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Event not found",
			"error":   err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(event)
}

func (e *EventRoutes) UpdateEvent(c *fiber.Ctx) error {
	id := c.Params("id")
	var event models.Event
	if err := e.DB.First(&event, "id = ?", id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Event not found",
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

	if err := e.DB.Model(&event).Updates(updateData).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update event",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(&fiber.Map{
		"message": "Event updated successfully",
		"event":   event,
	})
}
