package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
)



type Transaction struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// User & Product Info
	UserID      *uint  `gorm:"column:user_id;index" json:"user_id"`
	ProductID   *uint  `gorm:"column:product_id" json:"product_id"`
	ProductName *string `gorm:"column:product_name" json:"product_name"`
	ProductType *string `gorm:"column:product_type;index" json:"product_type"`
	CustomerNo  string  `gorm:"column:customer_no;index;not null" json:"customer_no"`
	BuyerSkuCode string `gorm:"column:buyer_sku_code;not null" json:"buyer_sku_code"`

	// Transaction IDs
	OrderID       string  `gorm:"column:order_id;unique;not null" json:"order_id"`
	TransactionID *string `gorm:"column:transaction_id;unique" json:"transaction_id"`

	// Payment Info
	GrossAmount   decimal.Decimal  `gorm:"column:gross_amount;not null" json:"gross_amount"`
	SellingPrice decimal.Decimal `gorm:"column:selling_price" json:"selling_price"`
	PurchasePrice decimal.Decimal `gorm:"column:purchase_price" json:"purchase_price"`

	PaymentType       *string `gorm:"column:payment_type" json:"payment_type"`
	PaymentMethodName *string `gorm:"column:payment_method_name" json:"payment_method_name"`
	MidtransResponse datatypes.JSON `gorm:"column:midtrans_response" json:"midtrans_response"`

	// Status
	PaymentStatus    string  `gorm:"column:payment_status;default:pending;index" json:"payment_status"`
	DigiflazzStatus  *string `gorm:"column:digiflazz_status;index" json:"digiflazz_status"`
	StatusMessage    *string `gorm:"column:status_message" json:"status_message"`

	// Product Specific
	RefID          *string  `gorm:"column:ref_id;index" json:"ref_id"`
	SerialNumber   *string  `gorm:"column:serial_number" json:"serial_number"`
	CustomerName   *string  `gorm:"column:customer_name" json:"customer_name"`
	MeterNo        *string  `gorm:"column:meter_no" json:"meter_no"`
	SubscriberID   *string  `gorm:"column:subscriber_id" json:"subscriber_id"`
	Kwh            *float64 `gorm:"column:kwh;type:decimal(10,2)" json:"kwh"`
	VoucherCode    *string  `gorm:"column:voucher_code" json:"voucher_code"`
	Note           *string  `gorm:"column:note;type:text" json:"note"`

	// URLs & Contact
	URL           *string `gorm:"column:url" json:"url"`
	DeeplinkGopay *string `gorm:"column:deeplink_gopay" json:"deeplink_gopay"`
	WaPembeli     string  `gorm:"column:wa_pembeli;not null" json:"wa_pembeli"`

	// Raw JSON Data
	DigiflazzRequest  datatypes.JSON `gorm:"column:digiflazz_request" json:"digiflazz_request"`
	DigiflazzResponse datatypes.JSON `gorm:"column:digiflazz_response" json:"digiflazz_response"`
	DigiflazzCallback datatypes.JSON `gorm:"column:digiflazz_callback" json:"digiflazz_callback"`
	DigiflazzFlag     *string        `gorm:"column:digiflazz_flag" json:"digiflazz_flag"`

	// Retry & Timing
	RetryAt          *time.Time `gorm:"column:retry_at" json:"retry_at"`
	RetryCount       int        `gorm:"column:retry_count;default:0" json:"retry_count"`
	LastErrorCode    *string    `gorm:"column:last_error_code;size:10" json:"last_error_code"`
	SaldoDebitedAt   *time.Time `gorm:"column:saldo_debited_at" json:"saldo_debited_at"`
	DigiflazzSentAt  *time.Time `gorm:"column:digiflazz_sent_at" json:"digiflazz_sent_at"`

	// Timestamps
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
