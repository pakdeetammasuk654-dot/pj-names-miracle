package router

import (
	"github.com/gin-gonic/gin"
	"go-backend/delivery/http/handler"
)

func SetupRouter(healthHandler *handler.HealthHandler, nameHandler *handler.NameHandler) *gin.Engine {
	r := gin.Default()

	// Serve avatars and static images statically
	r.Static("/avatars", "./avatars")
	r.Static("/buddhas", "./thai_buddha")
	r.Static("/images", "./images")

	// Base API route group
	api := r.Group("/api/v1")
	{
		api.GET("/health", healthHandler.CheckHealth)
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello World from Gin Clean Architecture with Pipeline Pattern!",
			})
		})

		// Name Search API (Top 100 Best Names)
		api.POST("/names/search", nameHandler.SearchTopNames)
		api.POST("/names/rerank", nameHandler.RerankTopNames)

		// Celebrities API
		api.GET("/celebrities", nameHandler.GetCelebrities)

		// Decode & Pair Detail API
		api.GET("/decode", nameHandler.DecodeName)
		api.GET("/pair-detail", nameHandler.GetPairDetail)
	}

	return r
}
