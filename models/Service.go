package models

import (
	"time"

	"gorm.io/gorm"
)

type Service struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Basic Information
	Name       string  `gorm:"column:name;size:255;not null" json:"name"`
	Slug       string  `gorm:"column:slug;size:255;uniqueIndex;not null" json:"slug"`
	Logo       *string `gorm:"column:logo;size:255" json:"logo"`
	LogoPublicID *string `gorm:"column:logo_public_id;size:255" json:"logo_public_id"`
	Icon       *string `gorm:"column:icon;size:255" json:"icon"`
	IconPublicID *string `gorm:"column:icon_public_id;size:255" json:"icon_public_id"`
	CategoryID uint    `gorm:"column:category_id;not null;index" json:"category_id"`
	Category   Category  `json:"category" gorm:"foreignKey:CategoryID"` //INI UNTUK PANGGIL RELASI PAKAI Category di Preload("Category")

	// Description
	Description *string `gorm:"column:description;type:text" json:"description"`
	HowToTopup  *string `gorm:"column:how_to_topup;type:text" json:"how_to_topup"`
	Notes       *string `gorm:"column:notes;type:text" json:"notes"`

	// Customer Number Format
	CustomerNoFormat string `gorm:"column:customer_no_format;type:enum('satu_input','dua_input');default:'satu_input'" json:"customer_no_format"`

	// Format Configuration
	ExampleFormat    *string `gorm:"column:example_format;type:text" json:"example_format"`
	Field1Label      string  `gorm:"column:field1_label;size:255;default:'User ID'" json:"field1_label"`
	Field1Placeholder string `gorm:"column:field1_placeholder;size:255;default:'Masukkan User ID'" json:"field1_placeholder"`
	Field2Label      *string `gorm:"column:field2_label;size:255" json:"field2_label"`
	Field2Placeholder *string `gorm:"column:field2_placeholder;size:255" json:"field2_placeholder"`

	// Flags
	IsActive  bool `gorm:"column:is_active;default:true;index" json:"is_active"`
	IsPopular bool `gorm:"column:is_popular;default:false;index" json:"is_popular"`
	ViewCount int  `gorm:"column:view_count;default:0" json:"view_count"`

	// Timestamps & Soft Delete
	CreatedAt time.Time      `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}
