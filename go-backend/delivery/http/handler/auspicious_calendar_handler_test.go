package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestAuspiciousCalendarMonthReturnsThaiLunarDays(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuspiciousCalendarHandler(nil)
	h.clock = func() time.Time { return time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC) }
	r := gin.New()
	r.GET("/calendar", h.Month)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/calendar?year=2026&month=7", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
	for _, expected := range []string{"\"month_name\":\"กรกฎาคม\"", "\"lunar_label\"", "\"แรม 14 ค่ำ เดือน 8\"", "\"is_wan_phra\""} {
		if !strings.Contains(w.Body.String(), expected) {
			t.Fatalf("response missing %s: %s", expected, w.Body.String())
		}
	}
}

func TestAuspiciousCalendarMonthRejectsInvalidMonth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/calendar", NewAuspiciousCalendarHandler(nil).Month)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/calendar?month=13", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAuspiciousTaksaEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := NewAuspiciousCalendarHandler(nil)
	h.clock = func() time.Time {
		return time.Date(2026, time.August, 28, 12, 0, 0, 0, time.FixedZone("Asia/Bangkok", 7*60*60))
	}
	r := gin.New()
	r.GET("/api/v1/auspicious/taksa", h.Taksa)
	r.POST("/api/v1/auspicious/taksa", h.Taksa)

	// 1. Test Today Default
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/api/v1/auspicious/taksa", nil))
	if w1.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w1.Code, http.StatusOK, w1.Body.String())
	}
	if !strings.Contains(w1.Body.String(), "\"is_today_default\":true") {
		t.Fatalf("expected is_today_default true: %s", w1.Body.String())
	}

	// 2. The selected target/search date must not affect current progressed age.
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/api/v1/auspicious/taksa?birth_date=1982-07-10&birth_time=02:00&target_date=2026-06-01", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w2.Code, http.StatusOK, w2.Body.String())
	}
	body := w2.Body.String()
	if !strings.Contains(body, "\"is_today_default\":false") || !strings.Contains(body, "\"birth_day_code\":\"friday\"") || !strings.Contains(body, "\"progressed_age\":45") || !strings.Contains(body, "\"hero_roles\"") {
		t.Fatalf("unexpected taksa result: %s", body)
	}
}
