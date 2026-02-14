package models

import (
	"time"
)

type ApplicationSetting struct {
	ID uint `gorm:"primaryKey" json:"id"`

	ApplicationName string `gorm:"column:application_name;size:255;not null" json:"application_name"`
	ApplicationFee  string `gorm:"column:application_fee;size:255;not null" json:"application_fee"` // nominal (string sesuai Laravel)

	Saldo float64 `gorm:"column:saldo;default:0" json:"saldo"`

	TermsCondition string `gorm:"column:terms_condition;type:longtext;not null" json:"terms_condition"`
	PrivacyPolicy  string `gorm:"column:privacy_policy;type:longtext;not null" json:"privacy_policy"`

	Logo string `gorm:"column:logo;size:255;not null" json:"logo"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
