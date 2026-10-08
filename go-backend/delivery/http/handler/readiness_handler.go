package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

type ReadinessHandler struct {
	db                 *sql.DB
	calibrationProfile domain.SoundCalibrationProfile
	calibrationWarm    bool
}

func NewReadinessHandler(db *sql.DB, profile domain.SoundCalibrationProfile, calibrationWarm bool) *ReadinessHandler {
	return &ReadinessHandler{db: db, calibrationProfile: profile, calibrationWarm: calibrationWarm}
}

func (h *ReadinessHandler) CheckReadiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	databaseReady := h.db != nil && h.db.PingContext(ctx) == nil
	ready := databaseReady && h.calibrationWarm
	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, gin.H{
		"error": !ready,
		"data": gin.H{
			"status":                     map[bool]string{true: "READY", false: "NOT_READY"}[ready],
			"api_version":                domain.APIVersion,
			"ranking_contract_version":   domain.RankingContractVersion,
			"database_ready":             databaseReady,
			"sound_calibration_warm":     h.calibrationWarm,
			"sound_calibration_approved": h.calibrationProfile.Approved,
			"sound_calibration_source":   h.calibrationProfile.Source,
		},
	})
}
