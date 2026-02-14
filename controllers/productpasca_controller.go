package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetProductPasca(p *gin.Context)  {
	var productPasca []models.ProductPasca
	err := config.DB.Where("seller_product_status = ? AND buyer_product_status = ?", true, true).Find(&productPasca).Error

	if err != nil {
		p.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengambil data",
		})
		return
	}

	p.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": productPasca,
	})
}