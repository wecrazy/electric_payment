package pltmhmodel

import (
	"electric_payment/config"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ConnectionType merepresentasikan jenis sambungan listrik PLN.
type ConnectionType string

const (
	ConnectionPrepaid  ConnectionType = "prabayar"
	ConnectionPostpaid ConnectionType = "pascabayar"
)

// ElectricityBase berisi data umum untuk layanan listrik prabayar dan pascabayar.
type ElectricityBase struct {
	gorm.Model
	CustomerID   string         `json:"customer_id" gorm:"type:varchar(100);uniqueIndex;not null"`
	MeterNumber  string         `json:"meter_number" gorm:"type:varchar(100);uniqueIndex;not null"`
	TariffCode   string         `json:"tariff_code"`
	PowerVA      int            `json:"power_va"`
	Connection   ConnectionType `json:"connection" gorm:"type:varchar(20);not null"`
	ActivatedAt  *time.Time     `json:"activated_at,omitempty"`
	LastModified *time.Time     `json:"last_modified,omitempty"`
}

// Prepaid merepresentasikan akun listrik PLN prabayar (token).
type Prepaid struct {
	ElectricityBase
	BalanceKWh   float64       `json:"balance_kwh" gorm:"default:0"`
	LastToken    string        `json:"last_token,omitempty"`
	LastTopUpAt  *time.Time    `json:"last_topup_at,omitempty"`
	TopUpHistory []TopUpRecord `json:"topup_history,omitempty" gorm:"foreignKey:PrepaidID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// TopUpRecord menyimpan detail pengisian token prabayar.
type TopUpRecord struct {
	gorm.Model
	PrepaidID uint      `json:"prepaid_id" gorm:"index;not null"`
	Token     string    `json:"token" gorm:"not null"`
	AmountRp  int64     `json:"amount_rp"`
	AddedKWh  float64   `json:"added_kwh"`
	TopUpAt   time.Time `json:"topup_at"`
	Vendor    string    `json:"vendor,omitempty"`
}

// Postpaid merepresentasikan akun listrik PLN pascabayar (tagihan bulanan).
type Postpaid struct {
	ElectricityBase
	BillingCycleStart  time.Time  `json:"billing_cycle_start"`
	BillingCycleEnd    time.Time  `json:"billing_cycle_end"`
	CurrentUsageKWh    float64    `json:"current_usage_kwh" gorm:"default:0"`
	OutstandingBalance int64      `json:"outstanding_balance" gorm:"default:0"`
	LastPaymentAt      *time.Time `json:"last_payment_at,omitempty"`
	IsDisconnected     bool       `json:"is_disconnected" gorm:"default:false"`
}

// --- Konstruktor bantu ---

func NewPrepaid(customerID, meterNumber, tariffCode string, powerVA int) *Prepaid {
	now := time.Now()
	return &Prepaid{
		ElectricityBase: ElectricityBase{
			CustomerID:   customerID,
			MeterNumber:  meterNumber,
			TariffCode:   tariffCode,
			PowerVA:      powerVA,
			Connection:   ConnectionPrepaid,
			ActivatedAt:  &now,
			LastModified: &now,
		},
		BalanceKWh:   0,
		TopUpHistory: nil,
	}
}

func NewPostpaid(customerID, meterNumber, tariffCode string, powerVA int, cycleStart, cycleEnd time.Time) *Postpaid {
	now := time.Now()
	return &Postpaid{
		ElectricityBase: ElectricityBase{
			CustomerID:   customerID,
			MeterNumber:  meterNumber,
			TariffCode:   tariffCode,
			PowerVA:      powerVA,
			Connection:   ConnectionPostpaid,
			ActivatedAt:  &now,
			LastModified: &now,
		},
		BillingCycleStart:  cycleStart,
		BillingCycleEnd:    cycleEnd,
		CurrentUsageKWh:    0,
		OutstandingBalance: 0,
	}
}

// --- Fungsi logika bisnis ---

func (p *Prepaid) TopUp(token string, amountRp int64, addedKWh float64, vendor string) {
	now := time.Now()
	p.BalanceKWh += addedKWh
	p.LastToken = token
	p.LastTopUpAt = &now
	p.LastModified = &now
	p.TopUpHistory = append([]TopUpRecord{{
		Token:    token,
		AmountRp: amountRp,
		AddedKWh: addedKWh,
		TopUpAt:  now,
		Vendor:   vendor,
	}}, p.TopUpHistory...)
	if len(p.TopUpHistory) > 10 {
		p.TopUpHistory = p.TopUpHistory[:10]
	}
}

func (p *Prepaid) Consume(kwh float64) error {
	if kwh < 0 {
		return fmt.Errorf("jumlah kWh harus positif")
	}
	if p.BalanceKWh < kwh {
		return fmt.Errorf("saldo tidak cukup: %.3f kWh tersedia, dibutuhkan %.3f kWh", p.BalanceKWh, kwh)
	}
	p.BalanceKWh -= kwh
	now := time.Now()
	p.LastModified = &now
	return nil
}

func (pp *Postpaid) CalculateBill(ratePerKWh float64) int64 {
	usageCost := int64(pp.CurrentUsageKWh * ratePerKWh)
	return pp.OutstandingBalance + usageCost
}

func (pp *Postpaid) Pay(amountRp int64) {
	now := time.Now()
	if amountRp <= 0 {
		return
	}
	pp.OutstandingBalance -= amountRp
	if pp.OutstandingBalance < 0 {
		pp.OutstandingBalance = 0
	}
	pp.LastPaymentAt = &now
	pp.LastModified = &now
	if pp.OutstandingBalance == 0 {
		pp.IsDisconnected = false
	}
}

func (pp *Postpaid) AddUsage(kwh float64) error {
	if kwh < 0 {
		return fmt.Errorf("pemakaian kWh tidak boleh negatif")
	}
	pp.CurrentUsageKWh += kwh
	now := time.Now()
	pp.LastModified = &now
	return nil
}

// CanMakePayment checks if customer is allowed to make payment
// Returns error if customer is disconnected
func (pp *Postpaid) CanMakePayment() error {
	if pp.IsDisconnected {
		return fmt.Errorf("sambungan listrik terputus, hubungi customer service untuk aktivasi kembali")
	}
	return nil
}

// HasOutstandingBill checks if customer has any bill to pay
func (pp *Postpaid) HasOutstandingBill() bool {
	return pp.OutstandingBalance > 0
}

// GetNoBillMessage returns a user-friendly message when customer has no outstanding bill
func (pp *Postpaid) GetNoBillMessage() string {
	currentMonth := time.Now().Format("January")
	currentYear := time.Now().Format("2006")
	return fmt.Sprintf(
		"Anda tidak memiliki tagihan yang perlu dibayar untuk bulan %s %s. "+
			"Terima kasih atas pembayaran Anda yang tepat waktu!",
		currentMonth, currentYear,
	)
}

// GetDisconnectionMessage returns a user-friendly message for disconnected customers
func (pp *Postpaid) GetDisconnectionMessage(supportPhone string) string {
	return fmt.Sprintf(
		"Maaf, sambungan listrik Anda telah diputus karena tunggakan pembayaran. "+
			"Untuk mengaktifkan kembali layanan, silakan hubungi customer service kami di %s",
		supportPhone,
	)
}

// Table name
// TableName untuk ElectricityBase
func (ElectricityBase) TableName() string {
	return config.GetConfig().PLTMHLembangPalesan.TbElectricityBase
}

// TableName untuk Prepaid
func (Prepaid) TableName() string {
	return config.GetConfig().PLTMHLembangPalesan.TbPrepaid
}

// TableName untuk TopUpRecord
func (TopUpRecord) TableName() string {
	return config.GetConfig().PLTMHLembangPalesan.TbTopupRecord
}

// TableName untuk Postpaid
func (Postpaid) TableName() string {
	return config.GetConfig().PLTMHLembangPalesan.TbPostpaid
}

// --- Fungsi Migrasi ---

// MigrateAllTables menjalankan migrasi tabel menggunakan GORM untuk MySQL.
func MigrateAllTables(db *gorm.DB) error {
	logrus.Info("menjalankan migrasi tabel plmth lembang palesan")
	return db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").AutoMigrate(
		&Prepaid{},
		&Postpaid{},
		&TopUpRecord{},
		&Transaction{},
	)
}

// TODO: use this below
// // --- Konfigurasi koneksi MySQL ---
// // Format DSN: username:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
// dsn := os.Getenv("MYSQL_DSN")
// if dsn == "" {
// dsn = "root:password@tcp(127.0.0.1:3306)/pln_db?charset=utf8mb4&parseTime=True&loc=Local"
// }

// // --- Koneksi ke MySQL ---
// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
// if err != nil {
// log.Fatalf("Gagal konek ke MySQL: %v", err)
// }

// fmt.Println("✅ Berhasil terhubung ke database MySQL")

// // --- Jalankan migrasi tabel ---
// if err := models.MigrateAllTables(db); err != nil {
// log.Fatalf("Gagal migrasi: %v", err)
// }

// fmt.Println("✅ Migrasi tabel PLN berhasil dijalankan!")

// // --- Contoh penggunaan sederhana ---

// // Membuat akun listrik prabayar baru
// prepaid := models.NewPrepaid("CUST-1001", "MTR-1001", "R1/1300", 1300)
// if err := db.Create(prepaid).Error; err != nil {
// log.Fatalf("Gagal membuat akun prabayar: %v", err)
// }

// fmt.Printf("Akun prabayar dibuat: %+v\n", prepaid)

// // Top-up token
// prepaid.TopUp("12345678901234567890", 200000, 50.0, "PLN Mobile")
// db.Save(prepaid)

// fmt.Println("Token berhasil diinput, saldo sekarang:", prepaid.BalanceKWh, "kWh")

// // Membuat akun listrik pascabayar baru
// postpaid := models.NewPostpaid(
// "CUST-2001",
// "MTR-2001",
// "R1/2200",
// 2200,
// time.Date(2025, 10, 1, 0, 0, 0, 0, time.UTC),
// time.Date(2025, 10, 31, 23, 59, 59, 0, time.UTC),
// )

// if err := db.Create(postpaid).Error; err != nil {
// fmt.Println("Pembayaran berhasil! Status pascabayar diperbarui.")
