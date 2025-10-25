package pltmhmodel

import (
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// SeedPrepaidCustomers seeds default test prepaid customers
func SeedPrepaidCustomers(db *gorm.DB) {
	var prepaidCount int64
	db.Model(&Prepaid{}).Count(&prepaidCount)
	if prepaidCount > 0 {
		logrus.Info("Prepaid customers already exist, skipping seeding")
		return
	}

	logrus.Info("Seeding default prepaid customers for testing")

	prepaidCustomers := []*Prepaid{
		NewPrepaid("CUST-PREP-001", "MTR-PREP-001", "R1/1300", 1300),
		NewPrepaid("CUST-PREP-002", "MTR-PREP-002", "R1/2200", 2200),
		NewPrepaid("CUST-PREP-003", "MTR-PREP-003", "R2/3500", 3500),
	}

	for _, customer := range prepaidCustomers {
		if err := db.Create(customer).Error; err != nil {
			logrus.Errorf("Failed to seed prepaid customer %s: %v", customer.CustomerID, err)
		}
	}

	// Add some top-up history for the first customer
	firstCustomer := prepaidCustomers[0]
	firstCustomer.TopUp("12345678901234567890", 100000, 25.5, "PLN Mobile")
	firstCustomer.TopUp("09876543210987654321", 200000, 51.0, "PLN Web")
	db.Save(firstCustomer)

	logrus.Info("Prepaid customers seeded successfully")
}

// SeedPostpaidCustomers seeds default test postpaid customers
func SeedPostpaidCustomers(db *gorm.DB) {
	var postpaidCount int64
	db.Model(&Postpaid{}).Count(&postpaidCount)
	if postpaidCount > 0 {
		logrus.Info("Postpaid customers already exist, skipping seeding")
		return
	}

	logrus.Info("Seeding default postpaid customers for testing")

	now := time.Now()
	cycleStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	cycleEnd := cycleStart.AddDate(0, 1, -1).Add(23 * time.Hour).Add(59 * time.Minute).Add(59 * time.Second)

	postpaidCustomers := []*Postpaid{
		NewPostpaid("CUST-POST-001", "MTR-POST-001", "R1/1300", 1300, cycleStart, cycleEnd),
		NewPostpaid("CUST-POST-002", "MTR-POST-002", "R1/2200", 2200, cycleStart, cycleEnd),
		NewPostpaid("CUST-POST-003", "MTR-POST-003", "R2/3500", 3500, cycleStart, cycleEnd),
	}

	for _, customer := range postpaidCustomers {
		if err := db.Create(customer).Error; err != nil {
			logrus.Errorf("Failed to seed postpaid customer %s: %v", customer.CustomerID, err)
		}
	}

	// Add some usage and payment for the first customer
	firstCustomer := postpaidCustomers[0]
	firstCustomer.AddUsage(150.5)
	firstCustomer.OutstandingBalance = 250000 // Simulate previous balance
	db.Save(firstCustomer)

	logrus.Info("Postpaid customers seeded successfully")
}

// SeedAllTestData seeds both prepaid and postpaid test customers
func SeedAllTestData(db *gorm.DB) {
	SeedPrepaidCustomers(db)
	SeedPostpaidCustomers(db)
}
