package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

type HealthHandler struct {
	healthUseCase domain.HealthUseCase
}

func NewHealthHandler(useCase domain.HealthUseCase) *HealthHandler {
	return &HealthHandler{
		healthUseCase: useCase,
	}
}

func (h *HealthHandler) CheckHealth(c *gin.Context) {
	clientIP := c.ClientIP()

	status, err := h.healthUseCase.CheckHealth(c.Request.Context(), clientIP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   false,
		"data":    status,
		"message": "Service is healthy",
	})
}
