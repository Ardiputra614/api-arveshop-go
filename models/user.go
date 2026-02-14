package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(255);not null" json:"name"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	EmailVerifiedAt *time.Time     `gorm:"type:timestamp;null" json:"email_verified_at"`
	Password        string         `gorm:"type:varchar(255);not null" json:"-"`
	RememberToken   *string        `gorm:"type:varchar(100)" json:"-"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}
