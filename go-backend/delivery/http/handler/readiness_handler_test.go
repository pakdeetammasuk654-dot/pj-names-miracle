package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

func TestReadinessFailsClosedWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/readiness", NewReadinessHandler(nil, domain.SoundCalibrationProfile{}, false).CheckReadiness)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/readiness", nil))
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), domain.RankingContractVersion) ||
		!strings.Contains(recorder.Body.String(), `"database_ready":false`) {
		t.Fatalf("readiness contract incomplete: %s", recorder.Body.String())
	}
}
