package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"api-arveshop-go/requests"
	"api-arveshop-go/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPaymentMethod(p *gin.Context)  {
	var PaymentMethod []models.PaymentMethod

	err  := config.DB.Where("is_active = ?", true).Find(&PaymentMethod).Error

	if err != nil {
		p.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengambil data",
		})

		return
	}

	p.JSON(http.StatusOK, gin.H{
		"message": "Berhasil mengambil data",
		"data": PaymentMethod,
	})
}


func CreatePaymentMethod(p *gin.Context) {
	var req requests.CreatePaymentMethod

	if err := p.ShouldBind(&req); err != nil {
        p.JSON(http.StatusBadRequest, gin.H{
            "message": "Data tidak valid",
            "error":   err.Error(),
        })
        return
    }

	var logoURL, logoPublicId string

	if req.Logo != nil {
        if err := utils.ValidateImage(req.Logo); err != nil {
            p.JSON(http.StatusBadRequest, gin.H{
                "message": "Logo tidak valid: " + err.Error(),
            })
            return
        }

        result, err := utils.UploadFile(req.Logo, "payment method/logo")
        if err != nil {
            p.JSON(http.StatusInternalServerError, gin.H{
                "message": "Gagal upload icon",
                "error":   err.Error(),
            })
            return
        }
        logoURL = result.SecureURL
		logoPublicId = result.PublicID
    }

	PaymentMethod := models.PaymentMethod{
		Name: req.Name,
		FeeType: req.FeeType,
		NominalFee: req.NominalFee,
		PercentaseFee: req.PercentaseFee,
		Type: req.Type,
		IsActive: req.IsActive,
		Logo: stringToPointer(logoURL),
		LogoPublicID: stringToPointer(logoPublicId),
	}

	err := config.DB.Create(&PaymentMethod).Error

	if err != nil {
		p.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal menambah data"})
		return
	}

	p.JSON(http.StatusCreated, gin.H{"message": "Berhasil menambah data", "data": PaymentMethod})
}