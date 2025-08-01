package routes

import (
	"Repos/ticketing-go/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

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

	// Basic input validation
	if req.UserID == 0 || req.EventID == 0 || req.TicketQty <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Missing or invalid user_id, event_id, or ticket_qty",
		})
	}

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// Lock the event row FOR UPDATE
		var event models.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", req.EventID).
			First(&event).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Event not found",
			})
		}

		// Validate event timing
		if event.EventDate.Before(time.Now()) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Cannot reserve ticket for past events",
			})
		}

		// Validate ticket quota
		if event.EventTicketAmount < req.TicketQty {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("Only %d tickets left", event.EventTicketAmount),
			})
		}

		// Deduct ticket amount
		event.EventTicketAmount -= req.TicketQty
		if err := tx.Save(&event).Error; err != nil {
			return fmt.Errorf("failed to update event quota: %w", err)
		}

		// Fetch merchant wallet from the event
		var merchant models.Merchant
		if err := tx.Where("id = ?", event.MerchantID).First(&merchant).Error; err != nil {
			return fmt.Errorf("failed to fetch merchant: %w", err)
		}

		// Declare priceUsd variable at function scope
		var priceUsd float64

		// Fetch and parse exchange rate from EXCHANGERATE API
		apiKey := os.Getenv("EXCHANGE_RATES_API_KEY")
		url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/pair/USD/IDR", apiKey)
		resp, err := http.Get(url)
		if err != nil {
			// Fallback to hardcoded rate if API request fails
			priceUsd = event.PriceIdr / 16500
		} else {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				priceUsd = event.PriceIdr / 16500
			} else {
				fmt.Println(string(bodyBytes)) // Print actual response content

				var rateData struct {
					Result         string  `json:"result"`
					ConversionRate float64 `json:"conversion_rate"`
				}

				if err := json.Unmarshal(bodyBytes, &rateData); err != nil {
					fmt.Println("Decode error:", err)
					priceUsd = event.PriceIdr / 16500
				} else if rateData.Result != "success" {
					fmt.Println("API returned non-success result")
					priceUsd = event.PriceIdr / 16500
				} else {
					priceUsd = event.PriceIdr / rateData.ConversionRate
				}
			}
		}
		// Build reservation
		reservation := models.Reservation{
			UserID:           req.UserID,
			EventID:          req.EventID,
			PaymentStatus:    "PENDING",
			PaymentDue:       time.Now().Add(1 * time.Hour),
			PaymentAmountUsd: float64(req.TicketQty) * priceUsd,       // assuming Event.PriceUsd exists
			PaymentAmountIdr: float64(req.TicketQty) * event.PriceIdr, // assuming Event.PriceIdr exists
			MerchantWallet:   merchant.MerchantWallet,
		}

		if err := tx.Create(&reservation).Error; err != nil {
			return fmt.Errorf("failed to create reservation: %w", err)
		}

		// Send response including merchant wallet
		c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":     "Reservation created",
			"reservation": reservation,
			// "wallet":      merchant.MerchantWallet,
			// "network" : "BSC",
			"payment": fiber.Map{
				"payment_address": merchant.MerchantWallet,
				"network":         "BSC",
				"amount_USDT":     float64(req.TicketQty) * priceUsd,
				"amount_IDR":      float64(req.TicketQty) * event.PriceIdr,
				"token":           "USDT",
				"token_address":   "0xCD60747D9Bbb1da2AfB2F834391f0FF6ccb15f1a",
				"token_standard":  "BEP20",
				"chain_id":        97, // BSC testnet
				"payment_due":     reservation.PaymentDue,
			},
		})
		return nil
	})

	if err != nil {
		if err, ok := err.(*fiber.Error); ok {
			return err
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Transaction failed",
			"error":   err.Error(),
		})
	}

	return nil
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
