package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID                string         `gorm:"primaryKey"`
	EventName         string         `json:"event_name"`
	EventDate         time.Time      `json:"event_date"`
	EventTime         time.Time      `json:"event_time"`
	EventVenue        string         `json:"event_venue"`
	EventGeolocation  string         `json:"event_geolocation"`
	EventDescription  string         `json:"event_description"`
	MerchantID        string         `json:"merchant_id"`
	EventTicketPrice  float64        `json:"event_ticket_price"`
	EventTicketAmount int            `json:"event_ticket_amount"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateEvents(db *gorm.DB) error {
	err := db.AutoMigrate(&Event{})
	return err
}
