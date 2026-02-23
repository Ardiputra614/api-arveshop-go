package routes

import (
	"api-arveshop-go/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.GET("/api/categories", controllers.GetCategoriesHome)
	r.GET("/api/services", controllers.GetServiceHome)
	r.GET("/api/products/:slug", controllers.GetProductHome)
	r.GET("/api/service/:slug", controllers.GetPersonalService)
	r.GET("/api/payment-method", controllers.GetPaymentMethodActive)
	r.POST("/api/create-transaction", controllers.CreateTransaction)
	r.POST("/api/get-products", controllers.GetProducts)

	api := r.Group("/api/admin")
	{
		api.GET("/users", controllers.GetUsers)
		api.POST("/users", controllers.CreateUser)
		api.GET("/application", controllers.GetApplicationSetting)
		
		api.GET("/categories", controllers.GetCategories)
		api.POST("/categories", controllers.CreateCategory)
		api.PUT("/categories/:id", controllers.UpdateCategory)
		api.DELETE("/categories/:id", controllers.DeleteCategory)

		api.GET("/services", controllers.GetServices)
		api.DELETE("/services/:id", controllers.DeleteService)
		api.POST("/services", controllers.CreateService)
		api.PATCH("/services/:id", controllers.UpdateService)
		
		
		api.GET("/payment-method", controllers.GetPaymentMethod)
		api.POST("/payment-method", controllers.CreatePaymentMethod)
		api.PUT("/payment-method/:id", controllers.UpdatePaymentMethod)
		api.DELETE("/payment-method/:id", controllers.DeletePaymentMethod)

		
		api.GET("/product-pasca", controllers.GetProductPasca)
	}
}
