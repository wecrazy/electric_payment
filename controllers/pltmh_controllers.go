package controllers

import (
	"electric_payment/config"
	"electric_payment/fun"
	"electric_payment/internal/gormdb"
	"electric_payment/model"
	pltmhmodel "electric_payment/model/pltmh_model"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetPaymentPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		// Retrieve cookies from the request
		cookies := c.Request.Cookies()

		// Check if the "credentials" cookie exists
		var credentialsCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "credentials" {
				credentialsCookie = cookie
				break
			}
		}

		parameters := gin.H{
			"APP_NAME":         config.GetConfig().App.Name,
			"APP_LOGO":         config.GetConfig().App.Logo,
			"APP_VERSION":      config.GetConfig().App.Version,
			"APP_VERSION_NO":   config.GetConfig().App.VersionNo,
			"APP_VERSION_CODE": config.GetConfig().App.VersionCode,
			"APP_VERSION_NAME": config.GetConfig().App.VersionName,
			"APP_TITLE":        config.GetConfig().App.Name,
			"GLOBAL_URL":       fun.GLOBAL_URL,
			"PLTMH_EMAIL":      config.GetConfig().PLTMHLembangPalesan.Email,
			"PLTMH_PHONE":      config.GetConfig().PLTMHLembangPalesan.Phone,
			"PLTMH_URL":        config.GetConfig().PLTMHLembangPalesan.PublicURL,
			"PLTMH_FACEBOOK":   config.GetConfig().PLTMHLembangPalesan.Facebook,
			"PLTMH_INSTAGRAM":  config.GetConfig().PLTMHLembangPalesan.Instagram,
			"PLTMH_YOUTUBE":    config.GetConfig().PLTMHLembangPalesan.Youtube,
			"PLTMH_TWITTER":    config.GetConfig().PLTMHLembangPalesan.Twitter,
			"TOKEN_OPTIONS":    config.GetConfig().PLTMHLembangPalesan.TopupOptions,
		}

		if credentialsCookie != nil {
			var admin model.Admin
			if err := db.Where("session = ?", credentialsCookie.Value).First(&admin).Error; err != nil {
				fun.ClearCookiesAndRedirect(c, cookies)
			}
		}

		c.HTML(http.StatusOK, "payment-page.html", parameters)
	}
}

// CheckCustomerData handles customer verification for prepaid/postpaid
func CheckCustomerData() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			CustomerInput string `json:"customer_input" binding:"required"`
			PaymentType   string `json:"payment_type" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request data",
				"error":   err.Error(),
			})
			return
		}

		// Check based on payment type
		if request.PaymentType == "prepaid" {
			var prepaid pltmhmodel.Prepaid
			// Check by meter_number or customer_id
			err := db.Where("meter_number = ? OR customer_id = ?", request.CustomerInput, request.CustomerInput).First(&prepaid).Error

			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Customer not found in prepaid database",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Prepaid customer found",
				"data": gin.H{
					"customer_id":   prepaid.CustomerID,
					"meter_number":  prepaid.MeterNumber,
					"tariff_code":   prepaid.TariffCode,
					"power_va":      prepaid.PowerVA,
					"connection":    prepaid.Connection,
					"balance_kwh":   prepaid.BalanceKWh,
					"last_token":    prepaid.LastToken,
					"last_topup_at": prepaid.LastTopUpAt,
				},
			})

		} else if request.PaymentType == "postpaid" {
			var postpaid pltmhmodel.Postpaid
			// Check by meter_number or customer_id
			err := db.Where("meter_number = ? OR customer_id = ?", request.CustomerInput, request.CustomerInput).First(&postpaid).Error

			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": "Customer not found in postpaid database",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Postpaid customer found",
				"data": gin.H{
					"customer_id":         postpaid.CustomerID,
					"meter_number":        postpaid.MeterNumber,
					"tariff_code":         postpaid.TariffCode,
					"power_va":            postpaid.PowerVA,
					"connection":          postpaid.Connection,
					"billing_cycle_start": postpaid.BillingCycleStart.Format("2006-01-02"),
					"billing_cycle_end":   postpaid.BillingCycleEnd.Format("2006-01-02"),
					"current_usage_kwh":   postpaid.CurrentUsageKWh,
					"outstanding_balance": postpaid.OutstandingBalance,
					"last_payment_at":     postpaid.LastPaymentAt,
					"is_disconnected":     postpaid.IsDisconnected,
				},
			})

		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid payment type. Must be 'prepaid' or 'postpaid'",
			})
		}
	}
}

// ProcessPayment handles payment processing
func ProcessPayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web

		var request struct {
			CustomerID    string `json:"customer_id" binding:"required"`
			PaymentType   string `json:"payment_type" binding:"required"`
			Amount        int64  `json:"amount" binding:"required"`
			PaymentMethod string `json:"payment_method" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid request data",
				"error":   err.Error(),
			})
			return
		}

		// Here you would implement the actual payment processing logic
		// For now, we'll simulate a successful payment

		if request.PaymentType == "prepaid" {
			// Handle prepaid token purchase
			var prepaid pltmhmodel.Prepaid
			if err := db.Where("customer_id = ?", request.CustomerID).First(&prepaid).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Prepaid customer not found",
				})
				return
			}

			// Generate token (simplified - in real implementation, you'd call PLN API)
			token := fmt.Sprintf("%020d", time.Now().Unix())
			addedKWh := float64(request.Amount) / 1500.0 // Approximate kWh calculation

			// Add top-up
			prepaid.TopUp(token, request.Amount, addedKWh, "Web Payment")

			// Save to database
			if err := db.Save(&prepaid).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to save payment record",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Prepaid payment successful",
				"data": gin.H{
					"token":       token,
					"amount":      request.Amount,
					"added_kwh":   addedKWh,
					"new_balance": prepaid.BalanceKWh,
				},
			})

		} else if request.PaymentType == "postpaid" {
			// Handle postpaid bill payment
			var postpaid pltmhmodel.Postpaid
			if err := db.Where("customer_id = ?", request.CustomerID).First(&postpaid).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Postpaid customer not found",
				})
				return
			}

			// Process payment
			postpaid.Pay(request.Amount)

			// Save to database
			if err := db.Save(&postpaid).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Failed to save payment record",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Postpaid payment successful",
				"data": gin.H{
					"paid_amount":       request.Amount,
					"remaining_balance": postpaid.OutstandingBalance,
					"payment_date":      postpaid.LastPaymentAt,
				},
			})
		}
	}
}
