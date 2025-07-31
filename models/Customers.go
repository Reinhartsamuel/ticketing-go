package models

import (
	"time"

	"gorm.io/gorm"
)

type Customer struct {
	ID        string `gorm:"primaryKey"`
	Code      string
	Name      string
	Email     string
	Address   string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func MigrateCustomers(db *gorm.DB) error {
	err := db.AutoMigrate(&Merchant{})
	return err
}
