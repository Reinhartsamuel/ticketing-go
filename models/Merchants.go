package models

import (
	"time"

	"gorm.io/gorm"
)

type Merchant struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	MerchantOwnerID uint           `json:"merchant_owner_id"`
	MerchantName    string         `json:"merchant_name"`
	MerchantWallet  string         `json:"merchant_wallet"`
	MerchantType    string         `json:"merchant_type"`
	Events          []Event        `gorm:"foreignKey:MerchantID"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index;column:deleted_at"`
}

func MigrateMerchants(db *gorm.DB) error {
	err := db.AutoMigrate(&Merchant{})
	return err
}
