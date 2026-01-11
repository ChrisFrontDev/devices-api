package http

import (
	"devices-api/docs"
	"devices-api/internal/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures all HTTP routes
func SetupRouter(deviceService *service.DeviceService) *gin.Engine {
	router := gin.Default()

	// Programmatically set swagger info (for dynamic host configuration)
	docs.SwaggerInfo.Title = "Devices API"
	docs.SwaggerInfo.Description = "A production-ready REST API for managing devices"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		deviceHandler := NewDeviceHandler(deviceService)

		devices := v1.Group("/devices")
		{
			devices.POST("", deviceHandler.CreateDevice)
			devices.GET("", deviceHandler.ListDevices)
			devices.GET("/:id", deviceHandler.GetDevice)
			devices.PUT("/:id", deviceHandler.UpdateDevice)
			devices.PATCH("/:id", deviceHandler.PartialUpdateDevice)
			devices.DELETE("/:id", deviceHandler.DeleteDevice)
		}
	}

	return router
}
