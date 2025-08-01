package models

import (
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID             uint           `json:"user_id"`
	EventID            uint           `json:"event_id"`
	PaymentStatus      string         `json:"payment_status" gorm:"default:PENDING"`
	PaymentAmountUsd   float64        `json:"payment_amount_usd" gorm:"default:0.0"`
	PaymentAmountIdr   float64        `json:"payment_amount_idr" gorm:"default:0.0"`
	PaymentDue         time.Time      `json:"payment_due"`
	PaymentConfirmedAt time.Time      `json:"payment_confirmed_at"`
	PaymentTxHash      string         `json:"payment_tx_hash"`
	PaymentTxUrl       string         `json:"payment_tx_url"`
	MerchantWallet     string         `json:"merchant_wallet"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateReservations(db *gorm.DB) error {
	err := db.AutoMigrate(&Reservation{})
	return err
}
