package tests

import (
	"electric_payment/config"
	"electric_payment/database"
	pltmhmodel "electric_payment/model/pltmh_model"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// go test -v -timeout 60m ./tests/pltmhLembangPalesan_test.go

func setupTestDB(t *testing.T) *gorm.DB {
	yamlCfg := config.GetConfig()
	db, err := database.InitAndCheckDB(
		yamlCfg.Database.Username,
		yamlCfg.Database.Password,
		yamlCfg.Database.Host,
		yamlCfg.Database.Port,
		yamlCfg.Database.Name,
	)

	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}

func TestPrepaidTopUp(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Get the first seeded prepaid customer
	var customer pltmhmodel.Prepaid
	err := db.Where("customer_id = ?", "CUST-PREP-001").First(&customer).Error
	assert.NoError(t, err)

	// Test initial balance
	assert.Equal(t, 0.0, customer.BalanceKWh)

	// Test top-up
	customer.TopUp("12345678901234567890", 100000, 25.5, "PLN Mobile")
	err = db.Save(&customer).Error
	assert.NoError(t, err)

	// Verify top-up
	assert.Equal(t, 25.5, customer.BalanceKWh)
	assert.Equal(t, "12345678901234567890", customer.LastToken)
	assert.NotNil(t, customer.LastTopUpAt)
	assert.Len(t, customer.TopUpHistory, 1)
	assert.Equal(t, "12345678901234567890", customer.TopUpHistory[0].Token)
	assert.Equal(t, int64(100000), customer.TopUpHistory[0].AmountRp)
	assert.Equal(t, 25.5, customer.TopUpHistory[0].AddedKWh)
}

func TestPrepaidConsume(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Get the first seeded prepaid customer
	var customer pltmhmodel.Prepaid
	err := db.Where("customer_id = ?", "CUST-PREP-001").First(&customer).Error
	assert.NoError(t, err)

	// Add balance first
	customer.TopUp("12345678901234567890", 100000, 50.0, "PLN Mobile")
	db.Save(&customer)

	// Test successful consumption
	err = customer.Consume(25.0)
	assert.NoError(t, err)
	assert.Equal(t, 25.0, customer.BalanceKWh)

	// Test insufficient balance
	err = customer.Consume(30.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "saldo tidak cukup")

	// Test negative consumption
	err = customer.Consume(-5.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "harus positif")
}

func TestPostpaidAddUsage(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Get the first seeded postpaid customer
	var customer pltmhmodel.Postpaid
	err := db.Where("customer_id = ?", "CUST-POST-001").First(&customer).Error
	assert.NoError(t, err)

	// Test initial usage
	assert.Equal(t, 0.0, customer.CurrentUsageKWh)

	// Test adding usage
	err = customer.AddUsage(150.5)
	assert.NoError(t, err)
	assert.Equal(t, 150.5, customer.CurrentUsageKWh)

	// Test negative usage
	err = customer.AddUsage(-10.0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak boleh negatif")
}

func TestPostpaidPay(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Get the first seeded postpaid customer
	var customer pltmhmodel.Postpaid
	err := db.Where("customer_id = ?", "CUST-POST-001").First(&customer).Error
	assert.NoError(t, err)

	// Set outstanding balance
	customer.OutstandingBalance = 250000
	db.Save(&customer)

	// Test payment
	customer.Pay(100000)
	assert.Equal(t, int64(150000), customer.OutstandingBalance)
	assert.NotNil(t, customer.LastPaymentAt)

	// Test overpayment
	customer.Pay(200000)
	assert.Equal(t, int64(0), customer.OutstandingBalance)

	// Test negative payment (should be ignored)
	customer.OutstandingBalance = 50000
	customer.Pay(-10000)
	assert.Equal(t, int64(50000), customer.OutstandingBalance)
}

func TestPostpaidCalculateBill(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Get the first seeded postpaid customer
	var customer pltmhmodel.Postpaid
	err := db.Where("customer_id = ?", "CUST-POST-001").First(&customer).Error
	assert.NoError(t, err)

	// Set usage and outstanding balance
	customer.CurrentUsageKWh = 200.0
	customer.OutstandingBalance = 50000
	db.Save(&customer)

	// Calculate bill with rate of 1500 per kWh
	ratePerKWh := 1500.0
	totalBill := customer.CalculateBill(ratePerKWh)

	// Expected: (200 * 1500) + 50000 = 300000 + 50000 = 350000
	expectedBill := int64(350000)
	assert.Equal(t, expectedBill, totalBill)
}

func TestPrepaidMultipleTopUps(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Get the second seeded prepaid customer
	var customer pltmhmodel.Prepaid
	err := db.Where("customer_id = ?", "CUST-PREP-002").First(&customer).Error
	assert.NoError(t, err)

	// Multiple top-ups
	customer.TopUp("11111111111111111111", 50000, 12.5, "PLN Web")
	customer.TopUp("22222222222222222222", 75000, 18.75, "PLN Mobile")
	customer.TopUp("33333333333333333333", 100000, 25.0, "PLN Office")

	// Save and reload to check top-up history
	db.Save(&customer)
	db.Preload("TopUpHistory").Where("customer_id = ?", "CUST-PREP-002").First(&customer)

	// Verify balance
	assert.Equal(t, 56.25, customer.BalanceKWh)

	// Verify top-up history (should be limited to 10, but we have 3)
	assert.Len(t, customer.TopUpHistory, 3)
	assert.Equal(t, "33333333333333333333", customer.TopUpHistory[0].Token) // Most recent first
	assert.Equal(t, "22222222222222222222", customer.TopUpHistory[1].Token)
	assert.Equal(t, "11111111111111111111", customer.TopUpHistory[2].Token)
}

func TestPostpaidBillingCycle(t *testing.T) {
	db := setupTestDB(t)
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Get the third seeded postpaid customer
	var customer pltmhmodel.Postpaid
	err := db.Where("customer_id = ?", "CUST-POST-003").First(&customer).Error
	assert.NoError(t, err)

	// Verify billing cycle dates
	now := time.Now()
	expectedStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	expectedEnd := expectedStart.AddDate(0, 1, -1).Add(23 * time.Hour).Add(59 * time.Minute).Add(59 * time.Second)

	assert.Equal(t, expectedStart, customer.BillingCycleStart)
	assert.Equal(t, expectedEnd, customer.BillingCycleEnd)
}
