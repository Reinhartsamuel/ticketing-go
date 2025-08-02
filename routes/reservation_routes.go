package routes

import (
	"Repos/ticketing-go/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
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
	app.Post("/reservations/manual-verification", r.ManualVerification)
}

func (r *ReservationRoutes) CreateReservation(c *fiber.Ctx) error {
	type ReservationItemRequest struct {
		EventID   uint `json:"event_id"`
		TicketQty uint `json:"ticket_qty"`
	}

	type ReservationRequest struct {
		UserID uint                     `json:"user_id"`
		Items  []ReservationItemRequest `json:"items"`
	}

	req := ReservationRequest{}
	if err := c.BodyParser(&req); err != nil || len(req.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Fetch exchange rate early
	apiKey := os.Getenv("EXCHANGE_RATES_API_KEY")
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/pair/USD/IDR", apiKey)

	exchangeRate := 16500.0 // fallback
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		var body struct {
			Result         string  `json:"result"`
			ConversionRate float64 `json:"conversion_rate"`
		}
		data, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(data, &body)
		if body.Result == "success" {
			exchangeRate = body.ConversionRate
		}
	}

	err = r.DB.Transaction(func(tx *gorm.DB) error {
		// Fetch customer
		var customer models.Customer
		if err := tx.First(&customer, req.UserID).Error; err != nil {
			return fmt.Errorf("customer not found")
		}

		var totalIDR, totalUSD float64
		var items []models.ReservationItem
		var merchantWallet string
		seenMerchants := make(map[uint]bool)

		for _, i := range req.Items {
			if i.EventID == 0 || i.TicketQty <= 0 {
				return fmt.Errorf("invalid event_id or ticket_qty")
			}

			var event models.Event
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", i.EventID).
				First(&event).Error; err != nil {
				return fmt.Errorf("event %d not found", i.EventID)
			}
			var merchant models.Merchant
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("id = ?", event.MerchantID).
				First(&merchant).Error; err != nil {
				return fmt.Errorf("merchant %d not found", event.MerchantID)
			}

			if event.EventDate.Before(time.Now()) {
				return fmt.Errorf("event %d has already occurred", i.EventID)
			}

			if event.EventTicketAmount < i.TicketQty {
				return fmt.Errorf("event %d only has %d tickets left", i.EventID, event.EventTicketAmount)
			}

			event.EventTicketAmount -= i.TicketQty
			if err := tx.Save(&event).Error; err != nil {
				return err
			}

			// Prevent mixing different merchants in a single reservation
			if _, ok := seenMerchants[event.MerchantID]; !ok && len(seenMerchants) > 0 {
				return fmt.Errorf("cannot reserve tickets from multiple merchants in one reservation")
			}
			seenMerchants[event.MerchantID] = true
			merchantWallet = merchant.MerchantWallet

			priceIDR := float64(i.TicketQty) * event.PriceIdr
			priceUSD := priceIDR / exchangeRate

			item := models.ReservationItem{
				EventID:      i.EventID,
				TicketTypeID: 0,
				Quantity:     uint(i.TicketQty),
				PriceIdr:     priceIDR,
				PriceUsd:     float64(int64(priceUSD*1000000)) / 1000000, // Limit to 6 decimal places for USDT precision
			}
			items = append(items, item)

			totalIDR += priceIDR
			totalUSD += float64(int64(priceUSD*1000000)) / 1000000 // Limit to 6 decimal places for USDT precision
		}

		// Create reservation
		reservation := models.Reservation{
			UserID:           req.UserID,
			PaymentStatus:    "PENDING",
			PaymentAmountUsd: totalUSD,
			PaymentAmountIdr: totalIDR,
			PaymentDue:       time.Now().Add(1 * time.Hour),
			MerchantWallet:   merchantWallet,
			CustomerWallet:   customer.WalletAddress,
		}

		if err := tx.Create(&reservation).Error; err != nil {
			return err
		}

		for i := range items {
			items[i].ReservationID = reservation.ID
		}
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
		var reservationWithItems models.Reservation
		if err := tx.Preload("ReservationItems").First(&reservationWithItems, reservation.ID).Error; err != nil {
			return err
		}
		c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message":     "Reservation created",
			"reservation": reservationWithItems,
			"payment": fiber.Map{
				"payment_address": reservation.MerchantWallet,
				"network":         "BSC",
				"amount_USDT":     float64(int64(totalUSD*1000000)) / 1000000, // Limit to 6 decimal places for USDT precision
				"amount_IDR":      totalIDR,
				"token":           "USDT",
				"token_address":   "0xCD60747D9Bbb1da2AfB2F834391f0FF6ccb15f1a",
				"token_standard":  "BEP20",
				"chain_id":        97,
				"payment_due":     reservation.PaymentDue,
			},
		})
		return nil
	})

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Reservation failed",
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

func (r *ReservationRoutes) ManualVerification(c *fiber.Ctx) error {
	type ManualVerificationRequest struct {
		ReservationID   uint   `json:"reservation_id"`
		TransactionHash string `json:"transaction_hash"`
	}

	req := ManualVerificationRequest{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}
	// create code for manual verficcation, user passes transaction hash,
	// we check to bsc testnet if the transaction is valid
	// and update reservation status accordingly
	reservation := models.Reservation{}
	if err := r.DB.First(&reservation, "id = ?", req.ReservationID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Reservation not found",
			"error":   err.Error(),
		})
	}

	url := "https://bnb-testnet.g.alchemy.com/v2/51MRDeFHeLtd5FrWrTMv0bsusLfs5n8r"
	// check if transaction hash is valid
	client, err := ethclient.Dial(url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to connect to RPC",
			"error":   err.Error(),
		})
	}
	defer client.Close()
	tx, err := client.TransactionReceipt(context.Background(), common.HexToHash(req.TransactionHash))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch transaction receipt",
			"error":   err.Error(),
		})
	}

	// check if transaction is valid
	if tx == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction hash",
			"error":   "Transaction not found",
		})
	}

	// check if transaction is from merchant wallet
	transaction, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(req.TransactionHash))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to fetch transaction",
			"error":   err.Error(),
		})
	}

	from, err := client.TransactionSender(context.Background(), transaction, tx.BlockHash, tx.TransactionIndex)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get transaction sender",
			"error":   err.Error(),
		})
	}

	if from != common.HexToAddress(reservation.MerchantWallet) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction sender",
			"error":   "Transaction sender is not the merchant wallet",
		})
	}

	// check if transaction is pending
	if isPending {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction",
			"error":   "Transaction is pending",
		})
	}

	// check if transaction is successful
	if tx.Status != 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction",
			"error":   "Transaction failed",
		})
	}

	// update reservation status
	reservation.PaymentStatus = "PAID"
	if err := r.DB.Save(&reservation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update reservation status",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Reservation verified successfully",
	})
}
