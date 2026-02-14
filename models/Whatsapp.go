package models

import (
	"time"	
)

type Whatsapp struct {
	ID uint `gorm:"primaryKey" json:"id"`

	Code   string `gorm:"column:code;size:255;not null" json:"code"`
	Name   string `gorm:"column:name;size:255;not null" json:"name"`
	Number string `gorm:"column:number;size:255;not null" json:"number"`

	QRCode string `gorm:"column:qr_code;type:text;not null" json:"qr_code"`
	Status string `gorm:"column:status;size:255;not null" json:"status"`

	MessagesSent   int `gorm:"column:messages_sent;default:0" json:"messages_sent"`
	MessagesFailed int `gorm:"column:messages_failed;default:0" json:"messages_failed"`

	LastActivity *string `gorm:"column:last_activity;size:255" json:"last_activity"`
	Uptime       string  `gorm:"column:uptime;size:255;not null" json:"uptime"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
