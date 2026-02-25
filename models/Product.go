package models

import (
	"fmt"
	"time"
)

type Product struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Informasi Dasar
	ProductName string `gorm:"column:product_name;size:255;not null" json:"product_name"`
	Slug        string `gorm:"column:slug;size:255;not null" json:"slug"`
	Category    string `gorm:"column:category;size:255;not null;index" json:"category"`
	Brand       string `gorm:"column:brand;size:255;not null" json:"brand"`
	Type        string `gorm:"column:type;size:255;not null" json:"type"`

	// prepaid / postpaid
	ProductType string `gorm:"column:product_type;size:50;not null;index" json:"product_type"`

	SellerName string `gorm:"column:seller_name;size:255;not null" json:"seller_name"`

	// Harga
	Price        int64 `gorm:"column:price;not null" json:"price"`
	SellingPrice int64 `gorm:"column:selling_price;not null" json:"selling_price"`

	// SKU dan Status
	BuyerSkuCode       string `gorm:"column:buyer_sku_code;size:255;not null;uniqueIndex" json:"buyer_sku_code"`
	BuyerProductStatus bool   `gorm:"column:buyer_product_status;not null;default:true" json:"buyer_product_status"`
	SellerProductStatus bool  `gorm:"column:seller_product_status;not null;default:true" json:"seller_product_status"`
	UnlimitedStock      bool  `gorm:"column:unlimited_stock;not null;default:false" json:"unlimited_stock"`
	Multi               bool  `gorm:"column:multi;not null;default:false" json:"multi"`

	// Stok (string karena Digiflazz kadang kirim "tersedia" / "habis")
	Stock string `gorm:"column:stock;size:50;not null;default:'0'" json:"stock"`

	// 🟢 CUTOFF TIME - Pakai string biasa
	StartCutOff string `gorm:"column:start_cut_off;size:5;not null;default:'00:00'" json:"start_cut_off"`
	EndCutOff   string `gorm:"column:end_cut_off;size:5;not null;default:'23:59'" json:"end_cut_off"`

	// Deskripsi
	Description string `gorm:"column:description;type:text" json:"desc"`

	// 🔴 TAMBAHKAN FIELD UNTUK JOB QUEUE
	Provider         string     `gorm:"column:provider;size:50;default:'digiflazz'" json:"provider"`
	LastSyncAt       *time.Time `gorm:"column:last_sync_at" json:"last_sync_at"`
	IsActive         bool       `gorm:"column:is_active;default:true" json:"is_active"`
	RetryCount       int        `gorm:"column:retry_count;default:0" json:"retry_count"`
	MaxRetry         int        `gorm:"column:max_retry;default:3" json:"max_retry"`
	RetryInterval    int        `gorm:"column:retry_interval;default:5" json:"retry_interval"` // menit

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName mengembalikan nama tabel yang benar
func (Product) TableName() string {
	return "products"
}

// IsWithinCutoff mengecek apakah waktu sekarang dalam cutoff
func (p *Product) IsWithinCutoff() bool {
	now := time.Now()
	currentTime := now.Format("15:04")
	
	start := p.StartCutOff
	end := p.EndCutOff
	
	// Jika tidak ada cutoff (default), return false
	if start == "00:00" && end == "23:59" {
		return false
	}
	
	// Case 1: Cutoff melewati tengah malam (start > end)
	if start > end {
		if currentTime >= start && currentTime <= "23:59" {
			return true
		}
		if currentTime >= "00:00" && currentTime <= end {
			return true
		}
	} else {
		// Case 2: Cutoff dalam hari yang sama
		if currentTime >= start && currentTime <= end {
			return true
		}
	}
	
	return false
}

// GetNextAvailableTime mengembalikan waktu berikutnya bisa diproses
func (p *Product) GetNextAvailableTime() *time.Time {
	if !p.IsWithinCutoff() {
		return nil
	}
	
	now := time.Now()
	currentTime := now.Format("15:04")
	end := p.EndCutOff
	
	var endHour, endMin int
	fmt.Sscanf(end, "%02d:%02d", &endHour, &endMin)
	
	nextTime := time.Date(
		now.Year(), now.Month(), now.Day(),
		endHour, endMin, 0, 0,
		now.Location(),
	)
	
	start := p.StartCutOff
	
	if start > end {
		if currentTime >= start || currentTime <= end {
			nextTime = nextTime.AddDate(0, 0, 1)
		}
	} else {
		if currentTime < end {
			// next time hari ini
		} else {
			var startHour, startMin int
			fmt.Sscanf(start, "%02d:%02d", &startHour, &startMin)
			nextTime = time.Date(
				now.Year(), now.Month(), now.Day()+1,
				startHour, startMin, 0, 0,
				now.Location(),
			)
		}
	}
	
	return &nextTime
}

// IsStockAvailable mengecek ketersediaan stok
func (p *Product) IsStockAvailable() bool {
	if p.UnlimitedStock {
		return true
	}
	return p.Stock != "0" && p.Stock != "habis" && p.Stock != ""
}

// CanBeProcessed mengecek apakah produk bisa diproses sekarang
func (p *Product) CanBeProcessed() bool {
	return p.IsActive && p.IsStockAvailable() && !p.IsWithinCutoff()
}

// GetTimeoutDuration mengembalikan timeout berdasarkan kategori
func (p *Product) GetTimeoutDuration() time.Duration {
	slowCategories := map[string]bool{
		"PLN": true, 
		"BPJS": true, 
		"TELKOM": true,
		"PASCABAYAR": true,
	}
	
	if slowCategories[p.Category] {
		return 60 * time.Second
	}
	return 30 * time.Second
}

// UpdateStock memperbarui stok produk
func (p *Product) UpdateStock(newStock string) {
	p.Stock = newStock
	p.UpdatedAt = time.Now()
}

// IsPrepaid mengecek apakah produk prepaid
func (p *Product) IsPrepaid() bool {
	return p.ProductType == "prepaid"
}

// IsPostpaid mengecek apakah produk postpaid
func (p *Product) IsPostpaid() bool {
	return p.ProductType == "postpaid"
}