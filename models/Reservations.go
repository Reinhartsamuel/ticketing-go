package models

import (
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID             uint           `json:"user_id"`
	PaymentStatus      string         `json:"payment_status" gorm:"default:PENDING"`
	PaymentAmountUsd   float64        `json:"payment_amount_usd" gorm:"default:0.0"`
	PaymentAmountIdr   float64        `json:"payment_amount_idr" gorm:"default:0.0"`
	PaymentDue         time.Time      `json:"payment_due"`
	PaymentConfirmedAt time.Time      `json:"payment_confirmed_at"`
	PaymentTxHash      string         `json:"payment_tx_hash"`
	PaymentTxUrl       string         `json:"payment_tx_url"`
	MerchantWallet     string         `json:"merchant_wallet"`
	CustomerWallet     string         `json:"customer_wallet"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index;column:deleted_at"`

	ReservationItems []ReservationItem `json:"reservation_items"`
}

// ReservationItem: represents each ticket purchased in the reservation
type ReservationItem struct {
	ID            uint    `gorm:"primaryKey;autoIncrement"`
	ReservationID uint    `json:"reservation_id"`            // This is the foreign key
	EventID       uint    `json:"event_id"`                  // in case each ticket is for a specific event
	TicketTypeID  uint    `json:"ticket_type_id"`            // optional if you support types (VIP, Reg, etc.)
	Quantity      uint    `json:"quantity" gorm:"default:1"` // usually 1, but allows batch purchase
	PriceUsd      float64 `json:"price_usd"`
	PriceIdr      float64 `json:"price_idr"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func MigrateReservations(db *gorm.DB) error {
	err := db.AutoMigrate(&Reservation{}, &ReservationItem{})
	return err
}
