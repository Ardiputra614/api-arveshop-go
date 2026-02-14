package models

import (
	"time"
)

type ProductPasca struct {
	ID uint `gorm:"primaryKey" json:"id"`

	ProductName string `gorm:"column:product_name;size:255;not null" json:"product_name"`
	Slug        string `gorm:"column:slug;size:255;not null;index" json:"slug"`
	Category    string `gorm:"column:category;size:255;not null;index" json:"category"`
	Brand       string `gorm:"column:brand;size:255;not null" json:"brand"`

	SellerName string `gorm:"column:seller_name;size:255;not null" json:"seller_name"`

	// Harga (string mengikuti Laravel)
	Price        string `gorm:"column:price;size:255;not null" json:"price"`          // harga beli (Digiflazz)
	SellingPrice string `gorm:"column:selling_price;size:255;not null" json:"selling_price"` // harga jual

	BuyerSkuCode        string `gorm:"column:buyer_sku_code;size:255;not null" json:"buyer_sku_code"`
	BuyerProductStatus  bool `gorm:"column:buyer_product_status; not null" json:"buyer_product_status"`
	SellerProductStatus bool   `gorm:"column:seller_product_status;not null" json:"seller_product_status"`

	StartCutOff string `gorm:"column:start_cut_off;size:255;not null" json:"start_cut_off"`
	EndCutOff   string `gorm:"column:end_cut_off;size:255;not null" json:"end_cut_off"`

	Admin      string `gorm:"column:admin;size:255;not null" json:"admin"`
	Commission string `gorm:"column:commission;size:255;not null" json:"commission"`
	Desc       string `gorm:"column:desc;size:255;not null" json:"desc"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}
