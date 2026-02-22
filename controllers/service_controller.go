package controllers

import (
	"log"
	"net/http"
	"strings"

	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"api-arveshop-go/requests"
	"api-arveshop-go/utils"

	"github.com/gin-gonic/gin"
)

func GetPersonalService(s *gin.Context) {
    slug := s.Param("slug")
    var service models.Service

    err := config.DB.Where("slug = ?", slug).First(&service).Error

    if err != nil {
        s.JSON(http.StatusNotFound, gin.H{"message": "Data tidak ada"})
        return
    }

    s.JSON(http.StatusOK, gin.H{"message":"Berhasil", "data": service})
}

func GetServiceHome(s *gin.Context) {
    var services []models.Service
    err := config.DB.Where("is_active = ?", true).Find(&services).Error
    if err != nil {
        s.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil data"})
        return
    }

    s.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": &services})
}


func GetServices(s *gin.Context) {
    var service []models.Service
    err := config.DB.
        Preload("Category").
        Find(&service).
        Error

    if err != nil {
        s.JSON(http.StatusInternalServerError, gin.H{
            "message": "Gagal Ambil Data",
            "error":   err.Error(),
        })
        return
    }

    s.JSON(http.StatusOK, gin.H{
        "message": "Berhasil",
        "data":    service,
    })
}

func GetServiceDetail(s *gin.Context) {
    id := s.Param("id")
    var service models.Service

    err := config.DB.
        Preload("Category").
        First(&service, id).
        Error

    if err != nil {
        s.JSON(http.StatusNotFound, gin.H{
            "message": "Service tidak ditemukan",
        })
        return
    }

    s.JSON(http.StatusOK, gin.H{
        "message": "Berhasil",
        "data":    service,
    })
}

func CreateService(s *gin.Context) {
    var req requests.CreateServiceRequest

    if err := s.ShouldBind(&req); err != nil {
        s.JSON(http.StatusBadRequest, gin.H{
            "message": "Data tidak valid",
            "error":   err.Error(),
        })
        return
    }

    var logoURL, iconPublicId, logoPublicId, iconURL string

    // Handle logo upload
    if req.Logo != nil {
        if err := utils.ValidateImage(req.Logo); err != nil {
            s.JSON(http.StatusBadRequest, gin.H{
                "message": "Logo tidak valid: " + err.Error(),
            })
            return
        }

        result, err := utils.UploadFile(req.Logo, "services/logos")
        if err != nil {
            s.JSON(http.StatusInternalServerError, gin.H{
                "message": "Gagal upload logo",
                "error":   err.Error(),
            })
            return
        }
        logoURL = result.SecureURL
		logoPublicId = result.PublicID
    }

    // Handle icon upload
    if req.Icon != nil {
        if err := utils.ValidateImage(req.Icon); err != nil {
            s.JSON(http.StatusBadRequest, gin.H{
                "message": "Icon tidak valid: " + err.Error(),
            })
            return
        }

        result, err := utils.UploadFile(req.Icon, "services/icons")
        if err != nil {
            s.JSON(http.StatusInternalServerError, gin.H{
                "message": "Gagal upload icon",
                "error":   err.Error(),
            })
            return
        }
        iconURL = result.SecureURL
		iconPublicId = result.PublicID
    }

    // Buat service model
    service := models.Service{
        Name:               req.Name,
        Slug:               generateSlug(req.Name),
        Logo:               stringToPointer(logoURL),
        Icon:               stringToPointer(iconURL),
		LogoPublicID: 		stringToPointer(logoPublicId),
		IconPublicID: 		stringToPointer(iconPublicId),
        CategoryID:         req.CategoryID,
        Description:        req.Description,
        HowToTopup:         req.HowToTopup,
        Notes:              req.Notes,
        CustomerNoFormat:   req.CustomerNoFormat,
        ExampleFormat:      req.ExampleFormat,
        Field1Label:        req.Field1Label,
        Field1Placeholder:  req.Field1Placeholder,
        Field2Label:        req.Field2Label,
        Field2Placeholder:  req.Field2Placeholder,
        IsActive:           req.IsActive,
        IsPopular:          req.IsPopular,
        ViewCount:          0,
    }

    // Simpan ke database
    if err := config.DB.Create(&service).Error; err != nil {
        s.JSON(http.StatusInternalServerError, gin.H{
            "message": "Gagal menambah data",
            "error":   err.Error(),
        })
        return
    }

    config.DB.Preload("Category").First(&service, service.ID)

    s.JSON(http.StatusCreated, gin.H{
        "message": "Berhasil menambah data",
        "data":    service,
    })
}

// Helper function untuk convert string ke *string
func stringToPointer(s string) *string {
    if s == "" {
        return nil
    }
    return &s
}

func UpdateService(s *gin.Context) {
    id := s.Param("id")
    var req requests.UpdateServiceRequest

    if err := s.ShouldBind(&req); err != nil {
        s.JSON(http.StatusBadRequest, gin.H{
            "message": "Data tidak valid",
            "error":   err.Error(),
        })
        return
    }

    // Cari service
    var service models.Service
    if err := config.DB.First(&service, id).Error; err != nil {
        s.JSON(http.StatusNotFound, gin.H{
            "message": "Service tidak ditemukan",
        })
        return
    }

    // Update fields
    if req.Name != "" {
        service.Name = req.Name
        if req.Slug == "" {
            service.Slug = generateSlug(req.Name)
        }
    }

    // ========== HANDLE LOGO ==========
    if req.RemoveLogo {
        // ✅ Hapus dari Cloudinary jika ada
        if service.LogoPublicID != nil && *service.LogoPublicID != "" {
            if err := utils.DeleteFile(*service.LogoPublicID); err != nil {
                log.Printf("Warning: Failed to delete logo: %v", err)
            }
        }
        service.Logo = nil
        service.LogoPublicID = nil
    } else if req.Logo != nil {
        // ✅ Hapus logo lama dari Cloudinary SEBELUM upload baru
        if service.LogoPublicID != nil && *service.LogoPublicID != "" {
            if err := utils.DeleteFile(*service.LogoPublicID); err != nil {
                log.Printf("Warning: Failed to delete old logo: %v", err)
            }
        }

        if err := utils.ValidateImage(req.Logo); err != nil {
            s.JSON(http.StatusBadRequest, gin.H{
                "message": "Logo tidak valid",
                "error":   err.Error(),
            })
            return
        }

        result, err := utils.UploadFile(req.Logo, "services/logos")
        if err != nil {
            s.JSON(http.StatusInternalServerError, gin.H{
                "message": "Gagal upload logo",
                "error":   err.Error(),
            })
            return
        }
        
        secureURL := result.SecureURL
        publicId := result.PublicID
        service.Logo = &secureURL
        service.LogoPublicID = &publicId
    }

    // ========== HANDLE ICON ==========
    if req.RemoveIcon {
        // ✅ Hapus dari Cloudinary jika ada
        if service.IconPublicID != nil && *service.IconPublicID != "" {
            if err := utils.DeleteFile(*service.IconPublicID); err != nil {
                log.Printf("Warning: Failed to delete icon: %v", err)
            }
        }
        service.Icon = nil
        service.IconPublicID = nil
    } else if req.Icon != nil {
        // ✅ Hapus icon lama dari Cloudinary SEBELUM upload baru
        if service.IconPublicID != nil && *service.IconPublicID != "" {
            if err := utils.DeleteFile(*service.IconPublicID); err != nil {
                log.Printf("Warning: Failed to delete old icon: %v", err)
            }
        }

        if err := utils.ValidateImage(req.Icon); err != nil {
            s.JSON(http.StatusBadRequest, gin.H{
                "message": "Icon tidak valid",
                "error":   err.Error(),
            })
            return
        }

        result, err := utils.UploadFile(req.Icon, "services/icons")
        if err != nil {
            s.JSON(http.StatusInternalServerError, gin.H{
                "message": "Gagal upload icon",
                "error":   err.Error(),
            })
            return
        }
        
        secureURL := result.SecureURL
        publicId := result.PublicID
        service.Icon = &secureURL
        service.IconPublicID = &publicId
    }

    // Update text fields
    if req.Description != nil {
        service.Description = req.Description
    }
    if req.HowToTopup != nil {
        service.HowToTopup = req.HowToTopup
    }
    if req.Notes != nil {
        service.Notes = req.Notes
    }
    if req.CustomerNoFormat != "" {
        service.CustomerNoFormat = req.CustomerNoFormat
    }
    if req.ExampleFormat != nil {
        service.ExampleFormat = req.ExampleFormat
    }
    if req.Field1Label != "" {
        service.Field1Label = req.Field1Label
    }
    if req.Field1Placeholder != "" {
        service.Field1Placeholder = req.Field1Placeholder
    }
    if req.Field2Label != nil {
        service.Field2Label = req.Field2Label
    }
    if req.Field2Placeholder != nil {
        service.Field2Placeholder = req.Field2Placeholder
    }
    if req.IsActive != nil {
        service.IsActive = *req.IsActive
    }
    if req.IsPopular != nil {
        service.IsPopular = *req.IsPopular
    }
    if req.CategoryID != 0 {
        service.CategoryID = req.CategoryID
    }

    // Save perubahan
    if err := config.DB.Save(&service).Error; err != nil {
        s.JSON(http.StatusInternalServerError, gin.H{
            "message": "Gagal mengupdate data",
            "error":   err.Error(),
        })
        return
    }

    // Preload category untuk response
    config.DB.Preload("Category").First(&service, service.ID)

    s.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mengupdate data",
        "data":    service,
    })
}

func DeleteService(s *gin.Context) {
    id := s.Param("id")
    
    // Cari service
    var service models.Service
    if err := config.DB.First(&service, id).Error; err != nil {
        s.JSON(http.StatusNotFound, gin.H{
            "message": "Data tidak ditemukan",
        })
        return
    }

    // Hapus logo dari Cloudinary jika ada
    if service.LogoPublicID != nil && *service.LogoPublicID != "" {
        if err := utils.DeleteFile(*service.LogoPublicID); err != nil {
            log.Printf("Warning: Failed to delete logo from Cloudinary: %v", err)
        }
    }

    // Hapus icon dari Cloudinary jika ada
    if service.IconPublicID != nil && *service.IconPublicID != "" {
        if err := utils.DeleteFile(*service.IconPublicID); err != nil {
            log.Printf("Warning: Failed to delete icon from Cloudinary: %v", err)
        }
    }

    // Hapus dari database (soft delete)
    if err := config.DB.Delete(&service).Error; err != nil {
        s.JSON(http.StatusInternalServerError, gin.H{
            "message": "Gagal menghapus service",
            "error":   err.Error(),
        })
        return
    }

    s.JSON(http.StatusOK, gin.H{
        "message": "Data berhasil dihapus!",
        "data": gin.H{
            "id": id,
        },
    })
}

// Helper function untuk generate slug
func generateSlug(name string) string {
    // Convert to lowercase
    slug := strings.ToLower(name)
    
    // Replace spaces with hyphens
    slug = strings.ReplaceAll(slug, " ", "-")
    
    // Remove special characters
    replacer := strings.NewReplacer(
        "'", "", "\"", "",
        ",", "", ".", "",
        "!", "", "?", "",
        "(", "", ")", "",
        "[", "", "]", "",
        "{", "", "}", "",
        "@", "", "#", "",
        "$", "", "%", "",
        "^", "", "&", "",
        "*", "", "+", "",
        "=", "", "~", "",
        "`", "", "|", "",
        "\\", "", "/", "",
    )
    slug = replacer.Replace(slug)
    
    // Remove multiple hyphens
    for strings.Contains(slug, "--") {
        slug = strings.ReplaceAll(slug, "--", "-")
    }
    
    // Trim hyphens
    slug = strings.Trim(slug, "-")
    
    return slug
}