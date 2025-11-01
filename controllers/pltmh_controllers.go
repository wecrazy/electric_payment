package controllers

import (
	"electric_payment/config"
	"electric_payment/fun"
	"electric_payment/internal/gormdb"
	"electric_payment/model"
	pltmhmodel "electric_payment/model/pltmh_model"
	"fmt"
	"image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nfnt/resize"
	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
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
			"ADMIN_FEE":        config.GetConfig().PLTMHLembangPalesan.AdminFee,
			"TARIF_CODES":      config.GetConfig().PLTMHLembangPalesan.TarifCode,
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
		switch request.PaymentType {
		case "prepaid":
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

		case "postpaid":
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

			// Check if customer is disconnected
			if postpaid.IsDisconnected {
				supportPhone := config.GetConfig().PLTMHLembangPalesan.Phone
				c.JSON(http.StatusOK, gin.H{
					"success":         false,
					"is_disconnected": true,
					"message": fmt.Sprintf(
						"Maaf, sambungan listrik Anda telah diputus karena tunggakan pembayaran. "+
							"Untuk mengaktifkan kembali layanan, silakan hubungi customer service kami di %s",
						supportPhone,
					),
					"support_contact": supportPhone,
				})
				return
			}

			// Check if customer has outstanding balance
			if postpaid.OutstandingBalance <= 0 {
				currentMonth := time.Now().Format("January")
				currentYear := time.Now().Format("2006")
				c.JSON(http.StatusOK, gin.H{
					"success":       false,
					"no_bill":       true,
					"message":       fmt.Sprintf("Anda tidak memiliki tagihan yang perlu dibayar untuk bulan %s %s. Terima kasih atas pembayaran Anda yang tepat waktu!", currentMonth, currentYear),
					"current_month": currentMonth,
					"current_year":  currentYear,
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

		default:
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

		// Get meter number from customer
		var meterNumber string
		switch request.PaymentType {
		case "prepaid":
			var prepaid pltmhmodel.Prepaid
			if err := db.Where("customer_id = ?", request.CustomerID).First(&prepaid).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Prepaid customer not found",
				})
				return
			}
			meterNumber = prepaid.MeterNumber

		case "postpaid":
			var postpaid pltmhmodel.Postpaid
			if err := db.Where("customer_id = ?", request.CustomerID).First(&postpaid).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"success": false,
					"message": "Postpaid customer not found",
				})
				return
			}

			// Check if customer is disconnected
			if postpaid.IsDisconnected {
				supportPhone := config.GetConfig().PLTMHLembangPalesan.Phone
				c.JSON(http.StatusForbidden, gin.H{
					"success":         false,
					"is_disconnected": true,
					"message": fmt.Sprintf(
						"Maaf, sambungan listrik Anda telah diputus. "+
							"Untuk mengaktifkan kembali layanan dan melakukan pembayaran, "+
							"silakan hubungi customer service kami di %s",
						supportPhone,
					),
					"support_contact": supportPhone,
				})
				return
			}

			// Check if customer has outstanding balance
			if postpaid.OutstandingBalance <= 0 {
				currentMonth := time.Now().Format("January")
				currentYear := time.Now().Format("2006")
				c.JSON(http.StatusBadRequest, gin.H{
					"success":       false,
					"no_bill":       true,
					"message":       fmt.Sprintf("Anda tidak memiliki tagihan yang perlu dibayar untuk bulan %s %s.", currentMonth, currentYear),
					"current_month": currentMonth,
					"current_year":  currentYear,
				})
				return
			}

			// Validate payment amount matches outstanding balance
			if request.Amount != postpaid.OutstandingBalance {
				c.JSON(http.StatusBadRequest, gin.H{
					"success":         false,
					"message":         fmt.Sprintf("Jumlah pembayaran tidak sesuai. Tagihan Anda: Rp %d", postpaid.OutstandingBalance),
					"expected_amount": postpaid.OutstandingBalance,
				})
				return
			}

			meterNumber = postpaid.MeterNumber

		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid payment type. Must be 'prepaid' or 'postpaid'",
			})
			return
		}

		// Generate unique transaction ID and payment token
		transactionID := fmt.Sprintf("TRX-%s-%d", request.CustomerID, time.Now().Unix())
		paymentToken := fmt.Sprintf("%s-%d", fun.GenerateRandomHexaString(32), time.Now().UnixNano())

		// Calculate total
		adminFee := int64(2500)
		totalAmount := request.Amount + adminFee

		// Convert payment type to ConnectionType
		var paymentType pltmhmodel.ConnectionType
		switch request.PaymentType {
		case "prepaid":
			paymentType = pltmhmodel.ConnectionPrepaid // "prabayar"
		case "postpaid":
			paymentType = pltmhmodel.ConnectionPostpaid // "pascabayar"
		}

		// Create transaction record
		transaction := pltmhmodel.Transaction{
			TransactionID: transactionID,
			CustomerID:    request.CustomerID,
			MeterNumber:   meterNumber,
			PaymentType:   paymentType,
			PaymentMethod: pltmhmodel.PaymentMethod(request.PaymentMethod),
			Amount:        request.Amount,
			AdminFee:      adminFee,
			TotalAmount:   totalAmount,
			Status:        pltmhmodel.TransactionPending,
			PaymentToken:  paymentToken,
			InitiatedAt:   time.Now(),
			ExpiredAt:     time.Now().Add(15 * time.Minute), // 15 minutes expiration
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
		}

		// For prepaid, generate token
		if request.PaymentType == "prepaid" {
			token := fmt.Sprintf("%020d", time.Now().Unix())
			addedKWh := float64(request.Amount) / 1500.0
			transaction.Token = token
			transaction.AddedKWh = addedKWh
		}

		// Save transaction to database
		if err := db.Create(&transaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create transaction record",
				"error":   err.Error(),
			})
			return
		}

		// Generate QR code
		selectedMainDir, err := fun.FindValidDirectory([]string{
			"web/file/transactions",
			"../web/file/transactions",
			"../../web/file/transactions",
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "No valid report directory found",
				"error":   err.Error(),
			})
			return
		}

		fileReportDir := filepath.Join(selectedMainDir, time.Now().Format("2006-01-02"))
		if err := os.MkdirAll(fileReportDir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to create report directory",
				"error":   err.Error(),
			})
			return
		}

		qrFileName := fmt.Sprintf("%s_%s_qr.png", transactionID, request.PaymentType)
		qrImgPath := filepath.Join(fileReportDir, qrFileName)

		// Payment validation URL
		paymentURL := fmt.Sprintf("%s/payment/validate/%s", config.GetConfig().PLTMHLembangPalesan.PublicURL, paymentToken)
		logoPath := "web/assets/self/img/logo_web.jpeg"

		if err := generateQRCodeURLWithLogo(paymentURL, logoPath, qrImgPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to generate QR code",
				"error":   err.Error(),
			})
			return
		}

		// Update transaction with QR code URL
		qrCodeURL := fmt.Sprintf("/file/transactions/%s/%s", time.Now().Format("2006-01-02"), qrFileName)
		transaction.QRCodeURL = qrCodeURL
		db.Save(&transaction)

		// Return response with QR code and transaction details
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Payment initiated successfully. Please scan QR code to complete payment.",
			"data": gin.H{
				"transaction_id": transactionID,
				"payment_token":  paymentToken,
				"qr_code_url":    qrCodeURL,
				"payment_url":    paymentURL,
				"amount":         request.Amount,
				"admin_fee":      adminFee,
				"total_amount":   totalAmount,
				"expires_at":     transaction.ExpiredAt.Format("2006-01-02 15:04:05"),
				"status":         string(transaction.Status),
			},
		})
	}
}

func generateQRCodeURLWithLogo(url, logoPath, imgOutputPath string) error {
	qr, err := qrcode.New(url)
	if err != nil {
		return err
	}

	file, err := os.Open(logoPath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	// Resize to 100x100 to ensure it's less than 1/5 of QR code size
	resizedImg := resize.Resize(100, 100, img, resize.Lanczos3)

	options := []standard.ImageOption{
		standard.WithLogoImage(resizedImg),
	}
	writer, err := standard.New(imgOutputPath, options...)
	if err != nil {
		return err
	}
	defer writer.Close()
	if err = qr.Save(writer); err != nil {
		return err
	}

	return nil
}

// ValidatePayment handles payment validation from QR code scan
func ValidatePayment() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web
		paymentToken := c.Param("token")

		// Find transaction by payment token
		var transaction pltmhmodel.Transaction
		if err := db.Where("payment_token = ?", paymentToken).First(&transaction).Error; err != nil {
			c.HTML(http.StatusNotFound, "payment-validation.html", gin.H{
				"success": false,
				"message": "Transaksi tidak ditemukan atau token tidak valid",
				"title":   "Pembayaran Gagal",
			})
			return
		}

		// Check if already completed
		if transaction.Status == pltmhmodel.TransactionCompleted {
			c.HTML(http.StatusOK, "payment-validation.html", gin.H{
				"success": true,
				"message": "Pembayaran sudah berhasil diproses sebelumnya",
				"title":   "Pembayaran Berhasil",
				"data": gin.H{
					"transaction_id": transaction.TransactionID,
					"amount":         transaction.TotalAmount,
					"completed_at":   transaction.CompletedAt.Format("2006-01-02 15:04:05"),
				},
			})
			return
		}

		// Check if expired
		if transaction.IsExpired() {
			transaction.MarkAsExpired()
			db.Save(&transaction)
			c.HTML(http.StatusGone, "payment-validation.html", gin.H{
				"success": false,
				"message": "Transaksi telah kadaluarsa. Silakan buat transaksi baru.",
				"title":   "Pembayaran Kadaluarsa",
			})
			return
		}

		// Process the payment based on type
		switch transaction.PaymentType {
		case pltmhmodel.ConnectionPrepaid:
			var prepaid pltmhmodel.Prepaid
			if err := db.Where("customer_id = ?", transaction.CustomerID).First(&prepaid).Error; err != nil {
				transaction.MarkAsFailed()
				db.Save(&transaction)
				c.HTML(http.StatusInternalServerError, "payment-validation.html", gin.H{
					"success": false,
					"message": "Gagal memproses pembayaran: Data pelanggan tidak ditemukan",
					"title":   "Pembayaran Gagal",
				})
				return
			}

			// Add top-up
			prepaid.TopUp(transaction.Token, transaction.Amount, transaction.AddedKWh, "Web Payment - QR")
			if err := db.Save(&prepaid).Error; err != nil {
				transaction.MarkAsFailed()
				db.Save(&transaction)
				c.HTML(http.StatusInternalServerError, "payment-validation.html", gin.H{
					"success": false,
					"message": "Gagal menyimpan data pembayaran",
					"title":   "Pembayaran Gagal",
				})
				return
			}

			// Mark transaction as completed
			transaction.MarkAsCompleted()
			db.Save(&transaction)

			c.HTML(http.StatusOK, "payment-validation.html", gin.H{
				"success": true,
				"message": "Pembayaran berhasil! Token listrik Anda sudah ditambahkan.",
				"title":   "Pembayaran Berhasil",
				"data": gin.H{
					"transaction_id": transaction.TransactionID,
					"token":          transaction.Token,
					"amount":         transaction.TotalAmount,
					"added_kwh":      transaction.AddedKWh,
					"new_balance":    prepaid.BalanceKWh,
				},
			})

		case pltmhmodel.ConnectionPostpaid:
			var postpaid pltmhmodel.Postpaid
			if err := db.Where("customer_id = ?", transaction.CustomerID).First(&postpaid).Error; err != nil {
				transaction.MarkAsFailed()
				db.Save(&transaction)
				c.HTML(http.StatusInternalServerError, "payment-validation.html", gin.H{
					"success": false,
					"message": "Gagal memproses pembayaran: Data pelanggan tidak ditemukan",
					"title":   "Pembayaran Gagal",
				})
				return
			}

			// Process payment
			postpaid.Pay(transaction.Amount)
			if err := db.Save(&postpaid).Error; err != nil {
				transaction.MarkAsFailed()
				db.Save(&transaction)
				c.HTML(http.StatusInternalServerError, "payment-validation.html", gin.H{
					"success": false,
					"message": "Gagal menyimpan data pembayaran",
					"title":   "Pembayaran Gagal",
				})
				return
			}

			// Mark transaction as completed
			transaction.MarkAsCompleted()
			db.Save(&transaction)

			c.HTML(http.StatusOK, "payment-validation.html", gin.H{
				"success": true,
				"message": "Pembayaran tagihan listrik berhasil diproses!",
				"title":   "Pembayaran Berhasil",
				"data": gin.H{
					"transaction_id":    transaction.TransactionID,
					"paid_amount":       transaction.TotalAmount,
					"remaining_balance": postpaid.OutstandingBalance,
					"payment_date":      transaction.CompletedAt.Format("2006-01-02 15:04:05"),
				},
			})

		default:
			transaction.MarkAsFailed()
			db.Save(&transaction)
			c.HTML(http.StatusBadRequest, "payment-validation.html", gin.H{
				"success": false,
				"message": "Tipe pembayaran tidak valid",
				"title":   "Pembayaran Gagal",
			})
		}
	}
}

// CheckPaymentStatus checks the status of a payment transaction
func CheckPaymentStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := gormdb.Databases.Web
		transactionID := c.Param("transaction_id")

		var transaction pltmhmodel.Transaction
		if err := db.Where("transaction_id = ?", transactionID).First(&transaction).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Transaction not found",
			})
			return
		}

		// Check if expired
		if transaction.IsExpired() && transaction.Status == pltmhmodel.TransactionPending {
			transaction.MarkAsExpired()
			db.Save(&transaction)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"transaction_id": transaction.TransactionID,
				"status":         string(transaction.Status),
				"completed_at":   transaction.CompletedAt,
				"expired_at":     transaction.ExpiredAt,
			},
		})
	}
}
