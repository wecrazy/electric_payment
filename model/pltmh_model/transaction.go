package pltmhmodel

import (
	"electric_payment/config"
	"time"

	"gorm.io/gorm"
)

// TransactionStatus represents the status of a payment transaction
type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionCompleted TransactionStatus = "completed"
	TransactionFailed    TransactionStatus = "failed"
	TransactionExpired   TransactionStatus = "expired"
)

// PaymentMethod represents the payment method used
type PaymentMethod string

const (
	PaymentMethodQRIS    PaymentMethod = "qris"
	PaymentMethodBank    PaymentMethod = "bank"
	PaymentMethodEWallet PaymentMethod = "ewallet"
	PaymentMethodManual  PaymentMethod = "manual" // Manual entry by admin
)

// Transaction merepresentasikan record transaksi pembayaran listrik
type Transaction struct {
	gorm.Model
	TransactionID string            `json:"transaction_id" gorm:"type:varchar(100);uniqueIndex;not null"`
	CustomerID    string            `json:"customer_id" gorm:"type:varchar(100);index;not null"`
	MeterNumber   string            `json:"meter_number" gorm:"type:varchar(100);index;not null"`
	PaymentType   ConnectionType    `json:"payment_type" gorm:"type:varchar(20);not null"` // prepaid/postpaid
	PaymentMethod PaymentMethod     `json:"payment_method" gorm:"type:varchar(20);not null"`
	Amount        int64             `json:"amount" gorm:"not null"`
	AdminFee      int64             `json:"admin_fee" gorm:"default:2500"`
	TotalAmount   int64             `json:"total_amount" gorm:"not null"`
	Status        TransactionStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Token         string            `json:"token,omitempty" gorm:"type:varchar(255)"` // For prepaid
	AddedKWh      float64           `json:"added_kwh,omitempty"`                      // For prepaid
	QRCodeURL     string            `json:"qr_code_url,omitempty" gorm:"type:varchar(500)"`
	PaymentToken  string            `json:"payment_token" gorm:"type:varchar(100);uniqueIndex"` // Token for validation
	InitiatedAt   time.Time         `json:"initiated_at"`
	CompletedAt   *time.Time        `json:"completed_at,omitempty"`
	ExpiredAt     time.Time         `json:"expired_at"`
	IPAddress     string            `json:"ip_address,omitempty" gorm:"type:varchar(50)"`
	UserAgent     string            `json:"user_agent,omitempty" gorm:"type:varchar(500)"`
}

// TableName untuk Transaction
func (Transaction) TableName() string {
	return config.GetConfig().PLTMHLembangPalesan.TbTransactionHistory
}

// IsExpired checks if the transaction has expired
func (t *Transaction) IsExpired() bool {
	return time.Now().After(t.ExpiredAt) && t.Status == TransactionPending
}

// MarkAsCompleted marks the transaction as completed
func (t *Transaction) MarkAsCompleted() {
	now := time.Now()
	t.Status = TransactionCompleted
	t.CompletedAt = &now
}

// MarkAsFailed marks the transaction as failed
func (t *Transaction) MarkAsFailed() {
	t.Status = TransactionFailed
}

// MarkAsExpired marks the transaction as expired
func (t *Transaction) MarkAsExpired() {
	t.Status = TransactionExpired
}
