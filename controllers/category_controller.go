package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"api-arveshop-go/requests"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetCategoriesHome(c *gin.Context) {
    var categories []models.Category
    err := config.DB.Where("is_active = ?", true).Find(&categories).Error

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"message": "Gagal mengambil data"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": categories})
}


func GetCategories(c *gin.Context)  {
	var categories []models.Category

	err := config.DB.Find(&categories).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal mengambil data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"Message": "Berhasil mengambil data",
		"data": categories,
	});
}


func CreateCategory(c *gin.Context)  {
	var req requests.CreateCategoryRequest

	err := c.ShouldBindJSON(&req)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": 400,
			"message": "Data tidak valid",
		})

		return
	}

	var category = models.Category{
		Name: req.Name,
		IsActive: req.IsActive,
	}

	err = config.DB.Create(&category).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menambah category",
		})

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code": 201,
		"message": "Berhasil menambah kategori",
        "data": category,
	})
}

func UpdateCategory(c *gin.Context) {
    id := c.Param("id")

    var req requests.UpdateCategoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": "Data tidak valid",
            "error":   err.Error(),
        })
        return
    }

    // Validasi input
    if req.Name == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": "Nama kategori wajib diisi",
        })
        return
    }

    if req.IsActive == nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "message": "Status aktif wajib diisi",
        })
        return
    }

    var category models.Category
    result := config.DB.Where("id = ?", id).First(&category)
    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "message": "Kategori tidak ditemukan",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "message": "Gagal mengambil data",
                "error":   result.Error.Error(),
            })
        }
        return
    }

    // Update hanya field yang diizinkan
    updates := map[string]interface{}{
        "name":      req.Name,
        "is_active": *req.IsActive, // Dereference pointer
        "updated_at": time.Now(), // Tambahkan updated_at jika ada di model
    }

    // Gunakan Model dengan Where untuk update
    result = config.DB.Model(&models.Category{}).
        Where("id = ?", id).
        Updates(updates)

    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "message": "Gagal mengubah data",
            "error":   result.Error.Error(),
        })
        return
    }

    if result.RowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{
            "message": "Kategori tidak ditemukan",
        })
        return
    }

    // Ambil data yang sudah diupdate
    var updatedCategory models.Category
    config.DB.Where("id = ?", id).First(&updatedCategory)

    c.JSON(http.StatusOK, gin.H{
        "message": "Berhasil mengubah data kategori",
        "code":    200,
        "data":    updatedCategory,
    })
}


func DeleteCategory(c *gin.Context)  {
	id := c.Param("id")
	var deletedCategory models.Category
	
	err := config.DB.Where("id = ?", &id).Delete(&deletedCategory).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Gagal menghapus",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil menghapus",
		"data": deletedCategory,
	})


}