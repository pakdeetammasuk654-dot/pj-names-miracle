package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-backend/usecase"
)

type EntitlementHandler struct {
	service usecase.EntitlementService
}

func NewEntitlementHandler(service usecase.EntitlementService) *EntitlementHandler {
	return &EntitlementHandler{service: service}
}

func (h *EntitlementHandler) VerifyPurchase(c *gin.Context) {
	var req usecase.StorePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "Invalid purchase payload"})
		return
	}

	result, err := h.service.VerifyPurchase(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrEntitlementUnavailable):
			log.Printf("entitlement verification unavailable: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": true, "message": "Purchase verification is temporarily unavailable"})
		case errors.Is(err, usecase.ErrInvalidStorePurchase):
			c.JSON(http.StatusUnauthorized, gin.H{"error": true, "message": "Purchase is invalid or inactive"})
		default:
			log.Printf("entitlement verification failed: %v", err)
			c.JSON(http.StatusBadGateway, gin.H{"error": true, "message": "Unable to verify purchase with the store"})
		}
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"error":             false,
		"entitlement_token": result.Token,
		"product_id":        result.ProductID,
		"platform":          result.Platform,
		"expires_at":        result.ExpiresAt,
	})
}
