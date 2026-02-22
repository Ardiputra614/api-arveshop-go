package models

import "time"

type Product struct {
	ID uint `gorm:"primaryKey" json:"id"`

	ProductName string `gorm:"size:255;not null" json:"product_name"`
	Slug        string `gorm:"size:255;not null" json:"slug"`
	Category    string `gorm:"size:255;not null;index" json:"category"`
	Brand       string `gorm:"size:255;not null" json:"brand"`
	Type        string `gorm:"size:255;not null" json:"type"`

	// prepaid / postpaid
	ProductType string `gorm:"size:50;not null;index" json:"product_type"`

	SellerName string `gorm:"size:255;not null" json:"seller_name"`

	// Harga tetap string (ikut Laravel / Digiflazz format)
	Price        int64 `gorm:"not null" json:"price"`
	SellingPrice int64 `gorm:"not null" json:"selling_price"`


	BuyerSkuCode       string `gorm:"size:255;not null;uniqueIndex" json:"buyer_sku_code"`
	BuyerProductStatus bool   `gorm:"not null;default:true" json:"buyer_product_status"`

	SellerProductStatus bool `gorm:"not null;default:true" json:"seller_product_status"`
	UnlimitedStock      bool `gorm:"not null;default:false" json:"unlimited_stock"`
	Multi               bool `gorm:"not null;default:false" json:"multi"`

	Stock string `gorm:"size:50;not null;default:'0'" json:"stock"`

	StartCutOff string `gorm:"type:time;not null;default:'00:00:00'"`
	EndCutOff   string `gorm:"type:time;not null;default:'23:59:59'"`



	Description string `gorm:"type:text;default:null" json:"desc"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
