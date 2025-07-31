package models

import (
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	ID string `gorm:"primaryKey"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateReservations(db *gorm.DB) error {
	err := db.AutoMigrate(&Reservation{})
	return err
}
