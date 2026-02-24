package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"net/http"

	"github.com/gin-gonic/gin"
)


func GetHistory(h *gin.Context) {
	order_id := h.Param("order_id")

	var history models.Transaction

	err := config.DB.Where("order_id = ?", &order_id).First(&history).Error

	if err != nil {
		h.JSON(http.StatusNotFound, gin.H{"code": 400, "message": "data tidak ditemukan"})
		return
	}

	h.JSON(http.StatusOK, gin.H{"message": "Berhasil", "data": history})
}