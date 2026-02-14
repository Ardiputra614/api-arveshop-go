package requests

import "mime/multipart"

type CreatePaymentMethod struct {
	Name string `form:"name" binding:"required"`
	FeeType string `form:"fee_type"`
	PercentaseFee float64 `form:"percentase_fee"`
	NominalFee float64 `form:"nominal_fee" `
	Type string `form:"type" binding:"required"`
	IsActive bool `form:"is_active"`
	Logo *multipart.FileHeader `form:"logo"`
	LogoPublicID *multipart.FileHeader `form:"logo_public_id"`
}