package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"api-arveshop-go/requests"
	"api-arveshop-go/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPaymentMethodActive(p *gin.Context) {
	var PaymentMethod []models.PaymentMethod
	err := config.DB.Where("is_active = ?", true).Find(&PaymentMethod).Error

	if err != nil {
		p.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil data"})
		return
	}

	if err != nil {
		p.JSON(http.StatusNotFound, gin.H{"message": "Data tidak ditemukan"})
		return
	}

	p.JSON(http.StatusOK, gin.H{"message": "berhasil", "data": &PaymentMethod})
}

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


func CreatePaymentMethod(c *gin.Context) {
	var req requests.CreatePaymentMethod

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Data tidak valid",
			"error":   err.Error(),
		})
		return
	}

	var logoURL, logoPublicId string

	// Handle logo upload
	if req.Logo != nil {
		if err := utils.ValidateImage(req.Logo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Logo tidak valid: " + err.Error(),
			})
			return
		}

		result, err := utils.UploadFile(req.Logo, "payment-methods/logos")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Gagal upload logo",
				"error":   err.Error(),
			})
			return
		}
		logoURL = result.SecureURL
		logoPublicId = result.PublicID
	}

	// Buat payment method model
	paymentMethod := models.PaymentMethod{
		Name:          req.Name,
		FeeType:       req.FeeType,
		NominalFee:    req.NominalFee,
		PercentaseFee: req.PercentaseFee,
		Type:          req.Type,
		IsActive:      req.IsActive,
		Logo:          logoURL,
		LogoPublicID:  logoPublicId,
	}

	// Simpan ke database
	if err := config.DB.Create(&paymentMethod).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menambah data",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Berhasil menambah data",
		"data":    paymentMethod,
	})
}

func UpdatePaymentMethod(p *gin.Context) {
    id := p.Param("id")
    var req requests.UpdatePaymentMethod

    // Bind form data
    if err := p.ShouldBind(&req); err != nil {
        p.JSON(http.StatusBadRequest, gin.H{
            "message": "Data tidak valid",
            "error":   err.Error(),
        })
        return
    }

    // Debug log untuk melihat apa yang terbinding
    log.Printf("Request: Name=%s, FeeType=%s, Type=%s, IsActive=%v, RemoveLogo=%v", 
        req.Name, req.FeeType, req.Type, req.IsActive, req.RemoveLogo)
    
    if req.Logo != nil {
        log.Printf("Logo file: %s, size: %d", req.Logo.Filename, req.Logo.Size)
    }

    // Cari data payment method yang akan diupdate
    var paymentMethod models.PaymentMethod
    err := config.DB.Where("id = ?", id).First(&paymentMethod).Error
    if err != nil {
        p.JSON(http.StatusNotFound, gin.H{"message": "Data tidak ditemukan"})
        return
    }

    // Update fields dari request
    paymentMethod.Name = req.Name
    paymentMethod.FeeType = req.FeeType
    paymentMethod.PercentaseFee = req.PercentaseFee
    paymentMethod.NominalFee = req.NominalFee
    paymentMethod.Type = req.Type
    paymentMethod.IsActive = req.IsActive

    // Handle logo
    if req.RemoveLogo {
        // Hapus logo dari Cloudinary jika ada
        if paymentMethod.LogoPublicID != "" {
            if err := utils.DeleteFile(paymentMethod.LogoPublicID); err != nil {
                log.Printf("Warning: Failed to delete logo: %v", err)
            }
        }
        paymentMethod.Logo = ""
        paymentMethod.LogoPublicID = ""
    } else if req.Logo != nil {
        // Ada file logo baru diupload
        log.Println("Processing new logo upload...")

        // Validasi file logo
        if err := utils.ValidateImage(req.Logo); err != nil {
            p.JSON(http.StatusBadRequest, gin.H{
                "message": "Logo tidak valid: " + err.Error(),
            })
            return
        }

        // Hapus logo lama dari Cloudinary jika ada
        if paymentMethod.LogoPublicID != "" {
            if err := utils.DeleteFile(paymentMethod.LogoPublicID); err != nil {
                log.Printf("Warning: Failed to delete old logo: %v", err)
            }
        }

        // Upload logo baru
        result, err := utils.UploadFile(req.Logo, "payment-methods/logos")
        if err != nil {
            p.JSON(http.StatusInternalServerError, gin.H{
                "message": "Gagal upload logo",
                "error":   err.Error(),
            })
            return
        }

        // Set logo baru
        paymentMethod.Logo = result.SecureURL
        paymentMethod.LogoPublicID = result.PublicID
        
        log.Printf("Logo uploaded successfully: %s", result.SecureURL)
    }

    // Simpan perubahan ke database
    if err := config.DB.Save(&paymentMethod).Error; err != nil {
        p.JSON(http.StatusInternalServerError, gin.H{
            "message": "Gagal mengupdate data",
            "error":   err.Error(),
        })
        return
    }

    // Preload category untuk response
    config.DB.Preload("Category").First(&paymentMethod, paymentMethod.ID)

    p.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mengupdate data",
        "data":    paymentMethod,
    })
}


func DeletePaymentMethod(p *gin.Context) {
	id := p.Param("id")

	var paymentMethod models.PaymentMethod
	err := config.DB.Where("id = ?", &id).First(&paymentMethod).Error

	if err != nil {
		p.JSON(http.StatusNotFound, gin.H{"message": "Data tidak ditemukan"})
		return
	}

	if paymentMethod.LogoPublicID != "" && paymentMethod.LogoPublicID != "" {
        if err := utils.DeleteFile(paymentMethod.LogoPublicID); err != nil {
            log.Printf("Warning: Failed to delete logo from Cloudinary: %v", err)
        }
    }

	// Hapus dari database (soft delete)
    if err := config.DB.Delete(&paymentMethod).Error; err != nil {
        p.JSON(http.StatusInternalServerError, gin.H{
            "message": "Gagal menghapus payment method",
            "error":   err.Error(),
        })
        return
    }

    p.JSON(http.StatusOK, gin.H{
        "message": "Data berhasil dihapus!",
        "data": gin.H{
            "id": id,
        },
    })


}