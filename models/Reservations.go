package models

import (
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint           `json:"user_id"`
	EventID         uint           `json:"event_id"`
	PaymentStatus   string         `json:"payment_status" gorm:"default:PENDING"`
	TransactionHash string         `json:"transaction_hash"`
	TransactionUrl  string         `json:"transaction_url"`
	BlockNumber     uint64         `json:"block_number"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateReservations(db *gorm.DB) error {
	err := db.AutoMigrate(&Reservation{})
	return err
}
