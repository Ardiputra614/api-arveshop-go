package controllers

import (
	"net/http"
	"api-arveshop-go/models"
	"api-arveshop-go/config"
	"github.com/gin-gonic/gin"
)

func GetApplicationSetting(a *gin.Context)  {
	var application []models.ApplicationSetting
	config.DB.Find(&application)
	a.JSON(http.StatusOK, application)
}
