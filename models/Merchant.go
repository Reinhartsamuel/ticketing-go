package models

import (
	"time"

	"gorm.io/gorm"
)

type Merchant struct {
	ID             uint   `gorm:"primarykey;autoIncrement:true" json:"id"`
	MerchantName   string `json:"merchant_name"`
	MerchantWallet string `json:"merchant_wallet"`
	MerchantType   string `json:"merchant_type"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func MigrateMerchants(db *gorm.DB) error {
	err := db.AutoMigrate(&Merchant{})
	return err
}
