package database

import (
	"electric_payment/model"
	whatsappmodel "electric_payment/model/whatsapp_model"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func AutoMigrateWeb(db *gorm.DB) {
	// Run migrations in stages to avoid foreign key issues

	// Stage 1: Create base models without foreign key relationships
	if err := db.AutoMigrate(
		&model.Admin{},
		&model.AdminStatus{},
		&model.AdminPasswordChangeLog{},
		&model.Role{},
		&model.RolePrivilege{},
		&model.Feature{},
		&model.LogActivity{},
		&model.Language{},
		&model.BadWord{},
		&model.AppConfig{},

		&model.WAMessage{},
		&model.WAMessageReply{},
		&model.WAPhoneUser{},
	); err != nil {
		logrus.Fatalf("Error while trying to automigrate db stage 1: %v", err)
	}

	// Stage 2: Create WhatsApp models without complex relationships
	if err := db.AutoMigrate(
		&whatsappmodel.WAContactInfo{},  // No dependencies
		&whatsappmodel.WAConversation{}, // Depends on Admin (User)
	); err != nil {
		logrus.Fatalf("Error while trying to automigrate db stage 2: %v", err)
	}

	// Stage 3: Create models that depend on conversations
	if err := db.AutoMigrate(
		&whatsappmodel.WAChatMessage{},      // Depends on WAConversation
		&whatsappmodel.WAGroupParticipant{}, // Depends on WAConversation
	); err != nil {
		logrus.Fatalf("Error while trying to automigrate db stage 3: %v", err)
	}

	// Stage 4: Create models that depend on messages
	if err := db.AutoMigrate(
		&whatsappmodel.WAMediaFile{},             // Depends on WAChatMessage
		&whatsappmodel.WAMessageDeliveryStatus{}, // Depends on WAChatMessage
	); err != nil {
		logrus.Fatalf("Error while trying to automigrate db: %v", err)
	}

	// Seeder
	seedRoles(db)
	seedFeature(db)
	seedRolePrivilege(db)
	seedAdmin(db)
	seedAdminStatus(db)
	seedAdminChangePwdLog(db)
	seedWhatsappLanguage(db)
	seedWhatsappPhoneUser(db)
	seedBadWords(db)
	seedAppConfig(db)

	seedIndonesiaRegion(db)

	// Seed Whatsapp Examples
	// controllers.SeedWhatsappSampleData(db)

}
