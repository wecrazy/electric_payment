package controllers

import (
	"electric_payment/internal/gormdb"
	pltmhmodel "electric_payment/model/pltmh_model"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ===========================
// CUSTOMER MANAGEMENT
// ===========================

// GetPrepaidCustomersTable returns prepaid customers for DataTables
func GetPrepaidCustomersTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var prepaidCustomers []pltmhmodel.Prepaid
		var totalRecords int64

		if err := db.Model(&pltmhmodel.Prepaid{}).Count(&totalRecords).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal menghitung pelanggan prabayar",
			})
			return
		}

		// Get all prepaid customers with their top-up history
		if err := db.Preload("TopUpHistory").Order("created_at DESC").Find(&prepaidCustomers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal mengambil data pelanggan prabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"data":         prepaidCustomers,
			"recordsTotal": totalRecords,
		})
	}
}

// GetPostpaidCustomersTable returns postpaid customers for DataTables
func GetPostpaidCustomersTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var postpaidCustomers []pltmhmodel.Postpaid
		var totalRecords int64

		// Count total records
		db.Model(&pltmhmodel.Postpaid{}).Count(&totalRecords)

		// Get all postpaid customers
		if err := db.Order("created_at DESC").Find(&postpaidCustomers).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal mengambil data pelanggan pascabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"data":         postpaidCustomers,
			"recordsTotal": totalRecords,
		})
	}
}

// CreatePrepaidCustomer creates a new prepaid customer
func CreatePrepaidCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			CustomerID  string `json:"customer_id" binding:"required"`
			MeterNumber string `json:"meter_number" binding:"required"`
			TariffCode  string `json:"tariff_code" binding:"required"`
			PowerVA     int    `json:"power_va" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Data permintaan tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Check if customer already exists
		var existingCustomer pltmhmodel.Prepaid
		if err := db.Where("customer_id = ? OR meter_number = ?", request.CustomerID, request.MeterNumber).First(&existingCustomer).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "ID Pelanggan atau No. Meter sudah terdaftar",
			})
			return
		}

		// Create new prepaid customer
		prepaid := pltmhmodel.NewPrepaid(request.CustomerID, request.MeterNumber, request.TariffCode, request.PowerVA)

		if err := db.Create(&prepaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal membuat pelanggan prabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pelanggan prabayar berhasil dibuat",
			"data":    prepaid,
		})
	}
}

// CreatePostpaidCustomer creates a new postpaid customer
func CreatePostpaidCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			CustomerID        string `json:"customer_id" binding:"required"`
			MeterNumber       string `json:"meter_number" binding:"required"`
			TariffCode        string `json:"tariff_code" binding:"required"`
			PowerVA           int    `json:"power_va" binding:"required"`
			BillingCycleStart string `json:"billing_cycle_start" binding:"required"`
			BillingCycleEnd   string `json:"billing_cycle_end" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Data permintaan tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Check if customer already exists
		var existingCustomer pltmhmodel.Postpaid
		if err := db.Where("customer_id = ? OR meter_number = ?", request.CustomerID, request.MeterNumber).First(&existingCustomer).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": "ID Pelanggan atau No. Meter sudah terdaftar",
			})
			return
		}

		// Parse dates
		cycleStart, err := time.Parse("2006-01-02", request.BillingCycleStart)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Format tanggal awal siklus tagihan tidak valid. Gunakan YYYY-MM-DD",
			})
			return
		}

		cycleEnd, err := time.Parse("2006-01-02", request.BillingCycleEnd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Format tanggal akhir siklus tagihan tidak valid. Gunakan YYYY-MM-DD",
			})
			return
		}

		// Create new postpaid customer
		postpaid := pltmhmodel.NewPostpaid(request.CustomerID, request.MeterNumber, request.TariffCode, request.PowerVA, cycleStart, cycleEnd)

		if err := db.Create(&postpaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal membuat pelanggan pascabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pelanggan pascabayar berhasil dibuat",
			"data":    postpaid,
		})
	}
}

// UpdatePrepaidCustomer updates an existing prepaid customer
func UpdatePrepaidCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			ID          uint    `json:"id" binding:"required"`
			CustomerID  string  `json:"customer_id"`
			MeterNumber string  `json:"meter_number"`
			TariffCode  string  `json:"tariff_code"`
			PowerVA     int     `json:"power_va"`
			BalanceKWh  float64 `json:"balance_kwh"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Data permintaan tidak valid",
				"error":   err.Error(),
			})
			return
		}

		var prepaid pltmhmodel.Prepaid
		if err := db.First(&prepaid, request.ID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan prabayar tidak ditemukan",
			})
			return
		}

		// Update fields
		if request.CustomerID != "" {
			prepaid.CustomerID = request.CustomerID
		}
		if request.MeterNumber != "" {
			prepaid.MeterNumber = request.MeterNumber
		}
		if request.TariffCode != "" {
			prepaid.TariffCode = request.TariffCode
		}
		if request.PowerVA > 0 {
			prepaid.PowerVA = request.PowerVA
		}
		if request.BalanceKWh >= 0 {
			prepaid.BalanceKWh = request.BalanceKWh
		}

		now := time.Now()
		prepaid.LastModified = &now

		if err := db.Save(&prepaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal memperbarui pelanggan prabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pelanggan prabayar berhasil diperbarui",
			"data":    prepaid,
		})
	}
}

// UpdatePostpaidCustomer updates an existing postpaid customer
func UpdatePostpaidCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			ID                 uint    `json:"id" binding:"required"`
			CustomerID         string  `json:"customer_id"`
			MeterNumber        string  `json:"meter_number"`
			TariffCode         string  `json:"tariff_code"`
			PowerVA            int     `json:"power_va"`
			CurrentUsageKWh    float64 `json:"current_usage_kwh"`
			OutstandingBalance int64   `json:"outstanding_balance"`
			IsDisconnected     *bool   `json:"is_disconnected"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Data permintaan tidak valid",
				"error":   err.Error(),
			})
			return
		}

		var postpaid pltmhmodel.Postpaid
		if err := db.First(&postpaid, request.ID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan pascabayar tidak ditemukan",
			})
			return
		}

		// Update fields
		if request.CustomerID != "" {
			postpaid.CustomerID = request.CustomerID
		}
		if request.MeterNumber != "" {
			postpaid.MeterNumber = request.MeterNumber
		}
		if request.TariffCode != "" {
			postpaid.TariffCode = request.TariffCode
		}
		if request.PowerVA > 0 {
			postpaid.PowerVA = request.PowerVA
		}
		if request.CurrentUsageKWh >= 0 {
			postpaid.CurrentUsageKWh = request.CurrentUsageKWh
		}
		if request.OutstandingBalance >= 0 {
			postpaid.OutstandingBalance = request.OutstandingBalance
		}
		if request.IsDisconnected != nil {
			postpaid.IsDisconnected = *request.IsDisconnected
		}

		now := time.Now()
		postpaid.LastModified = &now

		if err := db.Save(&postpaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal memperbarui pelanggan pascabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pelanggan pascabayar berhasil diperbarui",
			"data":    postpaid,
		})
	}
}

// DeletePrepaidCustomer deletes a prepaid customer
func DeletePrepaidCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web
		id := c.Param("id")

		var prepaid pltmhmodel.Prepaid
		if err := db.First(&prepaid, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan prabayar tidak ditemukan",
			})
			return
		}

		if err := db.Delete(&prepaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal menghapus pelanggan prabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pelanggan prabayar berhasil dihapus",
		})
	}
}

// DeletePostpaidCustomer deletes a postpaid customer
func DeletePostpaidCustomer() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web
		id := c.Param("id")

		var postpaid pltmhmodel.Postpaid
		if err := db.First(&postpaid, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan pascabayar tidak ditemukan",
			})
			return
		}

		if err := db.Delete(&postpaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal menghapus pelanggan pascabayar",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pelanggan pascabayar berhasil dihapus",
		})
	}
}

// ===========================
// TRANSACTION MANAGEMENT
// ===========================

// GetTransactionsTable returns all transactions
func GetTransactionsTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var transactions []pltmhmodel.Transaction
		var totalRecords int64

		// Count total records
		db.Model(&pltmhmodel.Transaction{}).Count(&totalRecords)

		// Get all transactions
		if err := db.Order("initiated_at DESC").Find(&transactions).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal mengambil data transaksi",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"data":         transactions,
			"recordsTotal": totalRecords,
		})
	}
}

// ManualTopUpPrepaid manually adds top-up for prepaid customer
func ManualTopUpPrepaid() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			CustomerID string  `json:"customer_id" binding:"required"`
			AmountRp   int64   `json:"amount_rp" binding:"required"`
			AddedKWh   float64 `json:"added_kwh" binding:"required"`
			Vendor     string  `json:"vendor"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Data permintaan tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Find customer
		var prepaid pltmhmodel.Prepaid
		if err := db.Where("customer_id = ?", request.CustomerID).First(&prepaid).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan prabayar tidak ditemukan",
			})
			return
		}

		// Generate token
		token := fmt.Sprintf("%020d", time.Now().Unix())

		// Set default vendor
		vendor := request.Vendor
		if vendor == "" {
			vendor = "Manual Admin"
		}

		// Add top-up
		prepaid.TopUp(token, request.AmountRp, request.AddedKWh, vendor)

		if err := db.Save(&prepaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal menambahkan top-up",
				"error":   err.Error(),
			})
			return
		}

		// Create Transaction record for tracking
		now := time.Now()
		transaction := pltmhmodel.Transaction{
			TransactionID: fmt.Sprintf("TRX-MANUAL-%s-%d", request.CustomerID, now.Unix()),
			CustomerID:    request.CustomerID,
			MeterNumber:   prepaid.MeterNumber,
			PaymentType:   pltmhmodel.ConnectionPrepaid,
			PaymentMethod: pltmhmodel.PaymentMethodManual,
			Amount:        request.AmountRp,
			AdminFee:      0, // No admin fee for manual entry
			TotalAmount:   request.AmountRp,
			Status:        pltmhmodel.TransactionCompleted,
			Token:         token,
			AddedKWh:      request.AddedKWh,
			InitiatedAt:   now,
			CompletedAt:   &now,
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		if err := db.Create(&transaction).Error; err != nil {
			// Log error but don't fail the request (top-up already succeeded)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Top-up berhasil ditambahkan (pencatatan transaksi gagal)",
				"data": gin.H{
					"customer_id":    prepaid.CustomerID,
					"token":          token,
					"balance_kwh":    prepaid.BalanceKWh,
					"last_topup":     prepaid.LastTopUpAt,
					"amount_paid":    request.AmountRp,
					"kwh_added":      request.AddedKWh,
					"transaction_id": transaction.TransactionID,
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Top-up berhasil ditambahkan",
			"data": gin.H{
				"customer_id":    prepaid.CustomerID,
				"token":          token,
				"balance_kwh":    prepaid.BalanceKWh,
				"last_topup":     prepaid.LastTopUpAt,
				"amount_paid":    request.AmountRp,
				"kwh_added":      request.AddedKWh,
				"transaction_id": transaction.TransactionID,
			},
		})
	}
}

// ManualPaymentPostpaid manually records payment for postpaid customer
func ManualPaymentPostpaid() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			CustomerID string `json:"customer_id" binding:"required"`
			AmountRp   int64  `json:"amount_rp" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Data permintaan tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Find customer
		var postpaid pltmhmodel.Postpaid
		if err := db.Where("customer_id = ?", request.CustomerID).First(&postpaid).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan pascabayar tidak ditemukan",
			})
			return
		}

		// Record payment
		postpaid.Pay(request.AmountRp)

		if err := db.Save(&postpaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal mencatat pembayaran",
				"error":   err.Error(),
			})
			return
		}

		// Create Transaction record for tracking
		now := time.Now()
		transaction := pltmhmodel.Transaction{
			TransactionID: fmt.Sprintf("TRX-MANUAL-%s-%d", request.CustomerID, now.Unix()),
			CustomerID:    request.CustomerID,
			MeterNumber:   postpaid.MeterNumber,
			PaymentType:   pltmhmodel.ConnectionPostpaid,
			PaymentMethod: pltmhmodel.PaymentMethodManual,
			Amount:        request.AmountRp,
			AdminFee:      0, // No admin fee for manual entry
			TotalAmount:   request.AmountRp,
			Status:        pltmhmodel.TransactionCompleted,
			InitiatedAt:   now,
			CompletedAt:   &now,
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		if err := db.Create(&transaction).Error; err != nil {
			// Log error but don't fail the request (payment already succeeded)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Pembayaran berhasil dicatat (pencatatan transaksi gagal)",
				"data": gin.H{
					"customer_id":         postpaid.CustomerID,
					"amount_paid":         request.AmountRp,
					"outstanding_balance": postpaid.OutstandingBalance,
					"last_payment":        postpaid.LastPaymentAt,
					"is_disconnected":     postpaid.IsDisconnected,
					"transaction_id":      transaction.TransactionID,
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pembayaran berhasil dicatat",
			"data": gin.H{
				"customer_id":         postpaid.CustomerID,
				"amount_paid":         request.AmountRp,
				"outstanding_balance": postpaid.OutstandingBalance,
				"last_payment":        postpaid.LastPaymentAt,
				"is_disconnected":     postpaid.IsDisconnected,
				"transaction_id":      transaction.TransactionID,
			},
		})
	}
}

// AddUsagePostpaid manually adds usage for postpaid customer
func AddUsagePostpaid() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			CustomerID string  `json:"customer_id" binding:"required"`
			UsageKWh   float64 `json:"usage_kwh" binding:"required"`
			RatePerKWh float64 `json:"rate_per_kwh" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Data permintaan tidak valid",
				"error":   err.Error(),
			})
			return
		}

		// Find customer
		var postpaid pltmhmodel.Postpaid
		if err := db.Where("customer_id = ?", request.CustomerID).First(&postpaid).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan pascabayar tidak ditemukan",
			})
			return
		}

		// Add usage
		if err := postpaid.AddUsage(request.UsageKWh); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		// Calculate new bill
		bill := postpaid.CalculateBill(request.RatePerKWh)
		postpaid.OutstandingBalance = bill

		if err := db.Save(&postpaid).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Gagal menambahkan pemakaian",
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Pemakaian berhasil ditambahkan",
			"data": gin.H{
				"customer_id":         postpaid.CustomerID,
				"usage_kwh":           request.UsageKWh,
				"current_usage_kwh":   postpaid.CurrentUsageKWh,
				"outstanding_balance": postpaid.OutstandingBalance,
			},
		})
	}
}

// GetCustomerDetail returns detailed information for a customer
func GetCustomerDetail() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web
		customerType := c.Query("type") // "prepaid" or "postpaid"
		customerID := c.Param("customer_id")

		if customerType == "prepaid" {
			var prepaid pltmhmodel.Prepaid
			if err := db.Preload("TopUpHistory").Where("customer_id = ?", customerID).First(&prepaid).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Pelanggan prabayar tidak ditemukan",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    prepaid,
			})
		} else if customerType == "postpaid" {
			var postpaid pltmhmodel.Postpaid
			if err := db.Where("customer_id = ?", customerID).First(&postpaid).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Pelanggan pascabayar tidak ditemukan",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    postpaid,
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid customer type. Must be 'prepaid' or 'postpaid'",
			})
		}
	}
}

// GetTopUpHistory returns top-up history for a prepaid customer
func GetTopUpHistory() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web
		customerID := c.Param("customer_id")
		limitStr := c.DefaultQuery("limit", "20")

		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 20
		}

		// Find customer
		var prepaid pltmhmodel.Prepaid
		if err := db.Where("customer_id = ?", customerID).First(&prepaid).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Pelanggan prabayar tidak ditemukan",
			})
			return
		}

		// Get top-up history
		var topUpHistory []pltmhmodel.TopUpRecord
		db.Where("prepaid_id = ?", prepaid.ID).Order("top_up_at DESC").Limit(limit).Find(&topUpHistory)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"customer_id":   customerID,
				"balance_kwh":   prepaid.BalanceKWh,
				"topup_history": topUpHistory,
				"total_records": len(topUpHistory),
			},
		})
	}
}
