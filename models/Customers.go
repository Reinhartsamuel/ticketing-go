package models

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string         `json:"name"`
	Email         string         `json:"email"`
	WalletAddress string         `json:"wallet_address"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateCustomers(db *gorm.DB) error {
	err := db.AutoMigrate(&Customer{})
	return err
}
