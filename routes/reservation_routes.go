package routes

import (
	"Repos/ticketing-go/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	type ReservationRequest struct {
		UserID    uint `json:"user_id"`
		EventID   uint `json:"event_id"`
		TicketQty int  `json:"ticket_qty"`
	}

	req := ReservationRequest{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// validate input!!
	if req.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "User ID is required",
		})
	}
	if req.EventID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Event ID is required",
		})
	}
	if req.TicketQty <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Ticket quantity must be greater than 0",
		})
	}

	// Begin a transaction
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		event := models.Event{}

		// Lock the event row FOR UPDATE (row-level lock) =>>>>>prevent race condition
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", req.EventID).
			First(&event).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Event not found",
			})
		}

		// Check if there's enough ticket quota
		if event.EventTicketAmount < req.TicketQty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("Not enough tickets available. Only %d left.", event.EventTicketAmount),
			})
		}

		// Decrement quota ----- put it as hold in PENDING reservation
		event.EventTicketAmount -= req.TicketQty
		if err := tx.Save(&event).Error; err != nil {
			return fmt.Errorf("failed to update ticket quota: %w", err)
		}

		// Create reservation with PENDING status
		reservation := models.Reservation{
			UserID:        req.UserID,
			EventID:       req.EventID,
			PaymentStatus: "PENDING",
		}
		if err := tx.Create(&reservation).Error; err != nil {
			return fmt.Errorf("failed to create reservation: %w", err)
		}

		// Everything succeeded — commit transaction
		c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":     "Reservation created and ticket quota held",
			"reservation": reservation,
		})
		return nil
	})

	// If error came from tx rollback
	if err != nil {
		// Already responded from within transaction if it's Fiber-compatible
		if err, ok := err.(*fiber.Error); ok {
			return err
		}
		// Otherwise, return generic error
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Transaction failed",
			"error":   err.Error(),
		})
	}

	return nil // handled inside tx
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
