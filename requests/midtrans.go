package requests

type CreateTransactionRequest struct {
	ID             string  `json:"id" binding:"required"`
	ProductName    string  `json:"product_name" binding:"required"`
	BuyerSkuCode   string  `json:"buyer_sku_code"`
	CustomerNo     string  `json:"customer_no" binding:"required"`
	SellingPrice   float64 `json:"selling_price" binding:"required"`
	Fee            float64 `json:"fee"`
	PaymentMethod  string  `json:"payment_method" binding:"required"`
	WaPembeli      string  `json:"wa_pembeli"`
	ProductType    string  `json:"product_type"`
}
