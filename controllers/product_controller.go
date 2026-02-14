package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetProductHome(p *gin.Context) {
    var products []models.Product
    slug := p.Param("slug")    
    
    // Gunakan First untuk single record, Find untuk multiple
    err := config.DB.
        Where("seller_product_status = ?", true).
        Where("buyer_product_status = ?", true).
        Where("slug = ?", slug).
        First(&products).Error // Gunakan First karena slug biasanya unique

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            p.JSON(http.StatusNotFound, gin.H{"message": "Data tidak ditemukan"})
            return
        }
        p.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil data: " + err.Error()})
        return
    }

    p.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": products})
}