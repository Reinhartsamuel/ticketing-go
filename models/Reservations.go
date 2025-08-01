package models

import (
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	ID               uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint           `json:"user_id"`
	EventID          uint           `json:"event_id"`
	PaymentStatus    string         `json:"payment_status" gorm:"default:PENDING"`
	PaymentAmountUsd float64        `json:"payment_amount_usd" gorm:"default:0.0"`
	PaymentAmountIdr float64        `json:"payment_amount_idr" gorm:"default:0.0"`
	PaymentDue       time.Time      `json:"payment_due"`
	TransactionHash  string         `json:"transaction_hash" gorm:"default:''"`
	TransactionUrl   string         `json:"transaction_url" gorm:"default:''"`
	BlockNumber      uint64         `json:"block_number" gorm:"default:0"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateReservations(db *gorm.DB) error {
	err := db.AutoMigrate(&Reservation{})
	return err
}
