package models

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        string         `gorm:"primaryKey"`
	Code      string         `json:"code"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Address   string         `json:"address"`
	Phone     string         `json:"phone"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateCustomers(db *gorm.DB) error {
	err := db.AutoMigrate(&Customer{})
	return err
}
