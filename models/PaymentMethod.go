package models

import (
	"time"
)

type PaymentMethod struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Name           string  `gorm:"column:name;size:255;not null" json:"name"`
	NominalFee     float64 `gorm:"column:nominal_fee;size:255" json:"nominal_fee"`
	PercentaseFee  float64 `gorm:"column:percentase_fee;size:255" json:"percentase_fee"`
	FeeType 		string `gorm:"column:fee_type;size:255" json:"fee_type"`

	// cc | qris | bank_transfer | ewallet | cstore
	Type string `gorm:"column:type;size:20;not null;index" json:"type"`

	Logo string `gorm:"column:logo;size:255" json:"logo"`
	LogoPublicID string `gorm:"column:logo_public_id;size:255" json:"logo_public_id"`

	// true | false
	IsActive    bool `gorm:"column:is_active;default:true;index" json:"is_active"`	

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
