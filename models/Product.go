package models

import (
	"time"
)

type Product struct {
	ID uint `gorm:"primaryKey" json:"id"`

	ProductName string `gorm:"column:product_name;size:255;not null" json:"product_name"`
	Slug        string `gorm:"column:slug;size:255;not null;index" json:"slug"`
	Category    string `gorm:"column:category;size:255;not null;index" json:"category"`
	Brand       string `gorm:"column:brand;size:255;not null" json:"brand"`
	Type        string `gorm:"column:type;size:255;not null" json:"type"`

	ProductType string `gorm:"column:product_type;size:255;not null" json:"product_type"` // prepaid / postpaid

	SellerName string `gorm:"column:seller_name;size:255;not null" json:"seller_name"`

	// Harga tetap string (ikut Laravel)
	Price        string `gorm:"column:price;size:255;not null" json:"price"`
	SellingPrice string `gorm:"column:selling_price;size:255;not null" json:"selling_price"`

	BuyerSkuCode       string `gorm:"column:buyer_sku_code;size:255;not null" json:"buyer_sku_code"`
	BuyerProductStatus bool `gorm:"column:buyer_product_status;not null" json:"buyer_product_status"`

	SellerProductStatus bool `gorm:"column:seller_product_status;not null" json:"seller_product_status"`
	UnlimitedStock      bool `gorm:"column:unlimited_stock;not null" json:"unlimited_stock"`
	Multi               bool `gorm:"column:multi;not null" json:"multi"`

	Stock string `gorm:"column:stock;size:255;not null" json:"stock"`

	StartCutOff time.Time `gorm:"column:start_cut_off;type:time;not null" json:"start_cut_off"`
	EndCutOff   time.Time `gorm:"column:end_cut_off;type:time;not null" json:"end_cut_off"`

	Desc string `gorm:"column:desc;size:255;not null" json:"desc"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
