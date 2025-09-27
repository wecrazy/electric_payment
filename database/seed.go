package database

import (
	"electric_payment/config"
	"electric_payment/fun"
	"electric_payment/model"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RolePermission represents the permission level for a role
type RolePermission struct {
	Create int8
	Read   int8
	Update int8
	Delete int8
}

func seedAdminStatus(db *gorm.DB) {
	var adminStatusCount int64
	db.Model(&model.AdminStatus{}).Count(&adminStatusCount)
	if adminStatusCount == 0 {
		var adminStatuses = []model.AdminStatus{
			{
				ID:        1,
				Title:     "PENDING",
				ClassName: "badge bg-label-warning",
			},
			{
				ID:        2,
				Title:     "ACTIVE",
				ClassName: "badge bg-label-success",
			},
			{
				ID:        3,
				Title:     "INACTIVE",
				ClassName: "badge bg-label-secondary",
			},
		}

		// Perform batch insert
		db.Create(&adminStatuses)

		for _, adminStatus := range adminStatuses {
			// Access IDs after insert
			logrus.Println("Insert New data  with ID : ", adminStatus.ID)
		}
	}
}

func seedAdmin(db *gorm.DB) {
	var adminCount int64
	db.Model(&model.Admin{}).Count(&adminCount)
	if adminCount == 0 {
		// Find role IDs dynamically
		var superAdminRole model.Role
		if err := db.Where("role_name = ?", "Super Admin").First(&superAdminRole).Error; err != nil {
			logrus.Fatalf("Error while trying to fetch Super Admin role: %v", err)
		}

		var admins = []model.Admin{
			{
				Fullname:     "Super Admin 1",
				Username:     "superadmin1",
				Phone:        config.GetConfig().Whatsmeow.WaSuperUser,
				Email:        "admin@website.com",
				Password:     fun.GenerateSaltedPassword("Ro224171222#"),
				Type:         0,
				Role:         int(superAdminRole.ID),
				Status:       2,
				CreateBy:     0,
				UpdateBy:     0,
				LastLogin:    time.Now(),
				ProfileImage: "uploads/admin/1.jpg",
			},
			{
				Fullname:  "Super Admin 2",
				Username:  "superadmin2",
				Phone:     config.GetConfig().Whatsmeow.WaSupport,
				Email:     "admin2@website.com",
				Password:  fun.GenerateSaltedPassword("Ro224171222#"),
				Type:      0,
				Role:      int(superAdminRole.ID),
				Status:    2,
				CreateBy:  0,
				UpdateBy:  0,
				LastLogin: time.Now(),
			},
		}

		// Perform batch insert
		db.Create(&admins)

		for _, admin := range admins {
			// Access IDs after insert
			logrus.Println("Insert New Admin  with ID : ", admin.ID)
		}
	}
}

func seedAdminChangePwdLog(db *gorm.DB) {
	var adminPasswordChangelogCount int64
	db.Model(&model.AdminPasswordChangeLog{}).Count(&adminPasswordChangelogCount)
	if adminPasswordChangelogCount == 0 {
		var admin_password_changelogs []model.AdminPasswordChangeLog

		var admins []model.Admin
		db.Find(&admins)
		for _, admin := range admins {
			admin_password_changelogs = append(admin_password_changelogs, model.AdminPasswordChangeLog{Email: admin.Email, Password: admin.Password})
		}

		// Perform batch insert
		db.Create(&admin_password_changelogs)

		for _, admin_password_changelog := range admin_password_changelogs {
			// Access IDs after insert
			logrus.Println("Insert New admin_password_changelog  with ID : ", admin_password_changelog.ID)
		}
	}
}

func seedRoles(db *gorm.DB) {
	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)
	if roleCount == 0 {
		roles := []model.Role{
			{
				RoleName:  "Super Admin",
				CreatedBy: 0,
				Icon:      "fal fa-user-crown",
				ClassName: "bg-label-primary",
			},
		}

		// Perform batch insert
		db.Create(&roles)

		for _, role := range roles {
			// Access IDs after insert
			logrus.Println("Insert New Roles ID : ", role.ID)
		}
	}
}

// getRolePermissions determines the permissions for a role based on its name and feature path
func getRolePermissions(roleName, featurePath string) (RolePermission, bool) {
	roleNameLower := strings.ToLower(roleName)
	featurePathLower := strings.ToLower(featurePath)

	// Super Admin gets full access to everything
	if strings.Contains(roleNameLower, "super admin") {
		return RolePermission{Create: 1, Read: 1, Update: 1, Delete: 1}, true
	}

	// Default role gets access to general features only
	if featurePath == "" ||
		strings.HasPrefix(featurePathLower, "tab-dashboard") {
		// strings.HasPrefix(featurePathLower, "tab-user-profile") {
		return RolePermission{Create: 0, Read: 1, Update: 0, Delete: 0}, true
	}

	// No access to restricted features
	return RolePermission{}, false
}

func seedRolePrivilege(db *gorm.DB) {
	var countData int64
	db.Model(&model.RolePrivilege{}).Count(&countData)
	if countData == 0 {
		var roleWebs []model.Role
		if result := db.Find(&roleWebs); result.Error != nil {
			logrus.Fatalf("Error while trying to fetch roles: %v", result.Error)
		}

		var features []model.Feature
		if result := db.Find(&features); result.Error != nil {
			logrus.Fatalf("Error while trying to fetch features: %v", result.Error)
		}

		var rolePrivileges []model.RolePrivilege

		for _, roleWeb := range roleWebs {
			for _, feature := range features {
				permission, hasAccess := getRolePermissions(roleWeb.RoleName, feature.Path)

				if hasAccess {
					rolePrivileges = append(rolePrivileges, model.RolePrivilege{
						RoleID:    roleWeb.ID,
						FeatureID: feature.ID,
						Create:    permission.Create,
						Read:      permission.Read,
						Update:    permission.Update,
						Delete:    permission.Delete,
					})
				}
			}
		}

		// Perform batch insert
		db.Create(&rolePrivileges)
	}
}

func seedFeature(db *gorm.DB) {
	var featureCount int64
	db.Model(&model.Feature{}).Count(&featureCount)
	if featureCount == 0 {
		var maxOrder uint
		db.Model(&model.Feature{}).Select("COALESCE(MAX(menu_order), 0)").Scan(&maxOrder)

		features := []model.Feature{
			{
				ParentID: 0,
				Title:    "Dashboard",
				Path:     "tab-dashboard",
				Status:   1,
				Level:    0,
				Icon:     "fad fa-tachometer-alt-fast",
			},
			/*
				Whatsapp
			*/
			{
				ParentID: 0,
				Title:    "Whatsapp",
				Path:     "",
				Status:   1,
				Level:    0,
				Icon:     "fab fa-whatsapp-square",
			},
			{
				ParentID: 0,
				Title:    "Bot Whatsapp",
				Path:     "tab-whatsapp",
				Status:   1,
				Level:    1,
				// Level:    0,
				Icon: "fad fa-user-robot",
			},
			{
				ParentID: 0,
				Title:    "Whatsapp User Management",
				Path:     "tab-whatsapp-user-management",
				Status:   1,
				Level:    1,
				// Level:    0,
				Icon: "fad fa-user-cog",
			},
			{
				ParentID: 0,
				Title:    "Chat & Messages",
				Path:     "tab-whatsapp-conversation",
				Status:   1,
				Level:    1,
				// Level:    0,
				Icon: "fad fa-whatsapp",
			},
			{
				ParentID: 0,
				Title:    "App Configuration",
				Path:     "tab-app-config",
				Status:   1,
				Level:    0,
				Icon:     "fad fa-cogs",
			},
			{
				ParentID: 0,
				Title:    "System User & Roles",
				Path:     "tab-roles",
				Status:   1,
				Level:    0,
				Icon:     "fad fa-users",
			},
			{
				ParentID: 0,
				Title:    "System Log",
				Path:     "tab-system-log",
				Status:   1,
				Level:    0,
				Icon:     "fad fa-terminal",
			},
			{
				ParentID: 0,
				Title:    "Log Activity",
				Path:     "tab-activity-log",
				Status:   1,
				Level:    0,
				Icon:     "fad fa-money-check-edit",
			},
			{
				ParentID: 0,
				Title:    "User Profile",
				Path:     "tab-user-profile",
				Status:   1,
				Level:    0,
				Icon:     "fad fa-id-card-alt",
			},
		}

		for i := range features {
			maxOrder++
			features[i].MenuOrder = maxOrder
		}

		// Perform batch insert
		db.Create(&features)

		// Set parent-child relationships
		parents := []struct {
			Title         string
			ChildPrefixes []string
		}{
			{
				Title:         "Whatsapp",
				ChildPrefixes: []string{"tab-whatsapp"},
			},
		}
		for _, p := range parents {
			var parent model.Feature
			if err := db.Where("title = ?", p.Title).First(&parent).Error; err != nil {
				logrus.Errorf("⚠️ Failed to find parent feature '%s': %v", p.Title, err)
				continue
			}

			for _, prefix := range p.ChildPrefixes {
				res := db.Model(&model.Feature{}).
					Where("path LIKE ?", prefix+"%").
					Update("parent_id", parent.ID)

				if res.Error != nil {
					logrus.Errorf("⚠️ Failed to update children for parent '%s' with prefix '%s': %v", p.Title, prefix, res.Error)
				} else {
					logrus.Infof("✅ Updated %d children for parent '%s' with prefix '%s'", res.RowsAffected, p.Title, prefix)
				}
			}
		}

		for _, feature := range features {
			logrus.Println("✅ Inserted DB Feature ID:", feature.ID, "| Menu Order:", feature.MenuOrder)
		}
	}
}

func seedWhatsappLanguage(db *gorm.DB) {
	// WhatsApp Bot Language
	var languageCount int64
	db.Model(&model.Language{}).Count(&languageCount)
	if languageCount == 0 {
		languages := []model.Language{
			{
				Name: "Bahasa Indonesia",
				Code: "id",
			},
			{
				Name: "English",
				Code: "us",
			},
			// Add more languages as needed
		}

		db.Create(&languages)

		for _, language := range languages {
			logrus.Println("🏳 Insert New Language with ID:", language.ID)
		}
	}
}

func seedWhatsappPhoneUser(db *gorm.DB) {
	// Seed Whatsapp Bot User
	var waPhoneUserCount int64
	db.Model(&model.WAPhoneUser{}).Count(&waPhoneUserCount)

	allowedTypes := model.AllWAMessageTypes
	jsonBytes, err := json.Marshal(allowedTypes)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal allowed types: %v", err))
	}

	if waPhoneUserCount == 0 {
		waUsers := []model.WAPhoneUser{
			{
				FullName:      "Super Admin 1",
				Email:         "admin@website.com",
				PhoneNumber:   config.GetConfig().Whatsmeow.WaSuperUser,
				IsRegistered:  true,
				AllowedChats:  model.BothChat,
				AllowedTypes:  datatypes.JSON(jsonBytes),
				AllowedToCall: true,
				Description:   "Phone number of Whatsapp SUPER USER 1",
				IsBanned:      false,
				UserType:      model.WaBotSuperUser,
				MaxDailyQuota: 250,
				UserOf:        model.UserOfPLTMH,
			},
		}
		botUsedPhoneNumber := config.GetConfig().Whatsmeow.WaBotUsed
		for i, phone := range botUsedPhoneNumber {
			fullName := ""
			switch i {
			case 0:
				fullName = "Bot Whatsapp (Development)"
			case 1:
				fullName = "Bot Whatsapp (Production)"
			default:
				fullName = fmt.Sprintf("Bot Whatsapp %d", i+1)
			}

			waUsers = append(waUsers, model.WAPhoneUser{
				FullName:      fullName,
				Email:         fmt.Sprintf("bot_wa_website_%d@gmail.com", i+1),
				PhoneNumber:   phone,
				IsRegistered:  true,
				AllowedChats:  model.BothChat,
				AllowedTypes:  datatypes.JSON(jsonBytes),
				AllowedToCall: true,
				Description:   "Phone number of Whatsapp Bot User",
				IsBanned:      false,
				UserType:      model.WaBotSuperUser,
				MaxDailyQuota: 250,
				UserOf:        model.UserOfPLTMH,
			})
		}

		for _, waUser := range waUsers {
			// Check if user with this phone number already exists
			var existingUser model.WAPhoneUser
			if err := db.Where("phone_number = ?", waUser.PhoneNumber).First(&existingUser).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					// User doesn't exist, create new one
					if err := db.Create(&waUser).Error; err != nil {
						panic(fmt.Sprintf("Failed to seed WhatsApp user with phone %s: %v", waUser.PhoneNumber, err))
					}
					logrus.Printf("✅ Created new WhatsApp user: %s (%s)", waUser.FullName, waUser.PhoneNumber)
				} else {
					// Some other database error occurred
					panic(fmt.Sprintf("Error checking existing WhatsApp user with phone %s: %v", waUser.PhoneNumber, err))
				}
			} else {
				// User already exists, skip creation
				logrus.Printf("⚠️ WhatsApp user with phone %s already exists, skipping creation for %s as %s (%s)", waUser.PhoneNumber, waUser.FullName, waUser.Description, waUser.UserType)
			}
		}
	}
}

func seedBadWords(db *gorm.DB) {
	// Check if there are already bad words
	var count int64
	db.Model(&model.BadWord{}).Count(&count)

	if count == 0 {
		// Seed data
		badWords := []model.BadWord{
			// Indonesian (id)
			{Word: "anjing", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "bajingan", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "jancok", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "jancuk", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "bangsat", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "bodoh", Language: "id", Category: model.CategoryGeneral, IsEnabled: true},
			{Word: "kontol", Language: "id", Category: model.CategorySexual, IsEnabled: true},
			{Word: "memek", Language: "id", Category: model.CategorySexual, IsEnabled: true},
			{Word: "ngentot", Language: "id", Category: model.CategorySexual, IsEnabled: true},
			{Word: "goblok", Language: "id", Category: model.CategoryGeneral, IsEnabled: true},
			{Word: "tolol", Language: "id", Category: model.CategoryGeneral, IsEnabled: true},
			{Word: "tai", Language: "id", Category: model.CategoryGeneral, IsEnabled: true},
			{Word: "setan", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "babi", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "kampret", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "puki", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "cukimai", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "telaso", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "tailaso", Language: "id", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "coklat", Language: "id", Category: model.CategoryRasis, IsEnabled: false}, // example of disabled word

			// English (en)
			{Word: "bitch", Language: "en", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "fuck", Language: "en", Category: model.CategoryUmpatan, IsEnabled: true},
			{Word: "idiot", Language: "en", Category: model.CategoryGeneral, IsEnabled: true},
			{Word: "nigger", Language: "en", Category: model.CategoryGeneral, IsEnabled: true},
			{Word: "nigga", Language: "en", Category: model.CategoryGeneral, IsEnabled: true},
			{Word: "dick", Language: "en", Category: model.CategoryGeneral, IsEnabled: true},
		}

		if err := db.Create(&badWords).Error; err != nil {
			logrus.Error("Failed to seed bad words: ", err)
		} else {
			logrus.Info("✅ Seeded bad words successfully")
		}
	}
}

func seedAppConfig(db *gorm.DB) {
	var count int64
	db.Model(&model.AppConfig{}).Count(&count)

	if count == 0 {
		// Find role IDs dynamically
		var superAdminRole model.Role
		if err := db.Where("role_name = ?", "Super Admin").First(&superAdminRole).Error; err != nil {
			logrus.Errorf("Error while trying to fetch Super Admin role: %v", err)
			return
		}

		appConfigs := []model.AppConfig{
			{
				RoleID:      superAdminRole.ID,
				AppName:     "Payment for Electricity",
				AppLogo:     "/assets/self/img/logo_web.png",
				AppVersion:  "Beta",
				VersionNo:   "1",
				VersionCode: "0.0.0.1.2025.07.31",
				VersionName: "electric_payment",
				IsActive:    true,
				Description: "Web Admin for managing electricity payments and monitoring transactions. Designed for administrators to oversee and manage the payment system efficiently.",
			},
		}

		if err := db.Create(&appConfigs).Error; err != nil {
			logrus.Error("Failed to seed app configs: ", err)
		} else {
			logrus.Info("✅ Seeded app configs successfully")
		}
	}
}

func seedIndonesiaRegion(db *gorm.DB) {
	// Get the table name from config
	tableName := config.GetConfig().Database.TbIndonesiaRegion

	// Check if table exists
	if !tableExists(db, tableName) {
		logrus.Infof("Table '%s' does not exist. Creating table structure first...", tableName)

		// Step 1: Create the table structure using GORM AutoMigrate
		if err := createIndonesiaRegionTable(db); err != nil {
			logrus.Errorf("Failed to create table structure for indonesia region: %v", err)
			return
		}

		logrus.Infof("✅ Table structure for '%s' created successfully", tableName)

		// Step 2: Check if table has data
		var count int64
		if err := db.Table(tableName).Count(&count).Error; err != nil {
			logrus.Errorf("Failed to count records in table '%s': %v", tableName, err)
			return
		}

		if count == 0 {
			logrus.Infof("Importing data from SQL file into table '%s'...", tableName)
			// Step 3: Import data from SQL file (only INSERT statements)
			if err := importIndonesiaRegionData(db, config.GetConfig().Database.DumpedIndonesiaRegionSQL); err != nil {
				logrus.Errorf("Failed to import data for indonesia region: %v", err)
				return
			}
			logrus.Infof("✅ Successfully imported data into table '%s'", tableName)
		} else {
			logrus.Infof("Table '%s' already contains %d records, skipping data import", tableName, count)
		}
	}
}

// createIndonesiaRegionTable creates the indonesia_region table structure using GORM AutoMigrate
func createIndonesiaRegionTable(db *gorm.DB) error {
	// Create the table with custom table name from config
	tableName := config.GetConfig().Database.TbIndonesiaRegion

	// Set custom table name for migration
	err := db.Table(tableName).AutoMigrate(&model.IndonesiaRegion{})
	if err != nil {
		return fmt.Errorf("failed to create table structure: %v", err)
	}

	return nil
}

// importIndonesiaRegionData imports only the INSERT data from SQL file
func importIndonesiaRegionData(db *gorm.DB, filePath string) error {
	// Get the absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("SQL file does not exist: %s", absPath)
	}

	// Read the SQL file
	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open SQL file: %v", err)
	}
	defer file.Close()

	// Read all content
	content, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read SQL file: %v", err)
	}

	// Get table name from config
	tableName := config.GetConfig().Database.TbIndonesiaRegion

	// Split the content by semicolons to get individual SQL statements
	sqlStatements := strings.Split(string(content), ";")

	// Execute only INSERT statements
	for _, statement := range sqlStatements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		// Skip comments and MySQL-specific commands
		if strings.HasPrefix(statement, "/*") ||
			strings.HasPrefix(statement, "--") ||
			strings.HasPrefix(statement, "/*!") {
			continue
		}

		// Skip CREATE TABLE statements (we already created the table)
		if strings.Contains(strings.ToUpper(statement), "CREATE TABLE") {
			continue
		}

		// Only process INSERT statements
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(statement)), "INSERT") {
			// Replace the table name in INSERT statement to use config table name
			statement = strings.ReplaceAll(statement, "`indonesia_region`", "`"+tableName+"`")
			statement = strings.ReplaceAll(statement, "indonesia_region", tableName)

			if err := db.Exec(statement).Error; err != nil {
				logrus.Warnf("Warning executing INSERT statement: %v", err)
				// Continue with other inserts even if one fails
			}
		}
	}

	return nil
}

// tableExists checks if a table exists in the database
func tableExists(db *gorm.DB, tableName string) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?)"
	err := db.Raw(query, tableName).Scan(&exists).Error
	if err != nil {
		logrus.Errorf("Error checking if table '%s' exists: %v", tableName, err)
		return false
	}
	return exists
}
