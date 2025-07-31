package migrations

import (
	"Repos/ticketing-go/models"

	"gorm.io/gorm"
)

func MigrateAll(db *gorm.DB) error {
	if err := models.MigrateMerchants(db); err != nil {
		return err
	}
	if err := models.MigrateCustomers(db); err != nil {
		return err
	}
	if err := models.MigrateEvents(db); err != nil {
		return err
	}
	if err := models.MigrateReservations(db); err != nil {
		return err
	}
	// Add other model migrations here
	// if err := models.MigrateUsers(db); err != nil {
	// 	return err
	// }
	return nil
}
