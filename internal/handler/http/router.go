package http

import (
	"devices-api/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all HTTP routes
func SetupRouter(deviceService *service.DeviceService) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

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
