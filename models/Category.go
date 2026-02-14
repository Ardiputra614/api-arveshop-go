package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID uint `grom:"id" gorm:"primaryKey" json:"id"`

	// Basic Information
	Name  string `gorm:"column:name;size:255;not null" json:"name"`
		
	IsActive    bool `gorm:"column:is_active;default:true;index" json:"is_active"`	

	// Timestamps & Soft Delete
	CreatedAt time.Time      `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;index" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}
