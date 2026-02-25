package controllers

import (
	"api-arveshop-go/config"
	"api-arveshop-go/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetApplicationSetting(a *gin.Context)  {
	var application []models.ProfilAplikasi
	config.DB.Find(&application)
	a.JSON(http.StatusOK, application)
}
