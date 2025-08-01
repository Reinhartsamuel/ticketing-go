package models

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	ID                uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	EventName         string         `json:"event_name"`
	EventDate         time.Time      `json:"event_date"`
	EventTime         time.Time      `json:"event_time"`
	EventVenue        string         `json:"event_venue"`
	EventGeolocation  string         `json:"event_geolocation"`
	EventDescription  string         `json:"event_description"`
	MerchantID        uint           `json:"merchant_id"`
	PriceIdr          float64        `json:"price_idr"`
	EventTicketAmount int            `json:"event_ticket_amount"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateEvents(db *gorm.DB) error {
	err := db.AutoMigrate(&Event{})
	return err
}
