package fun

import (
	"electric_payment/model"

	"gorm.io/gorm"
)

func RemoveEmailSession(db *gorm.DB, email string) {
	updates := map[string]interface{}{
		"LastLogin": 0,
		"Session":   "",
	}
	// Perform the update
	db.Model(&model.Admin{}).Where("Email = ?", email).Updates(updates)
}
