package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/KranaxALT/thailunar"
	"github.com/gin-gonic/gin"
	"go-backend/domain"
	"go-backend/usecase"
)

// AuspiciousCalendarHandler exposes the factual calendar layer only. Rules that
// decide whether a day is auspicious for an activity will be added separately.
type AuspiciousCalendarHandler struct {
	clock   func() time.Time
	service *usecase.AuspiciousService
}

func NewAuspiciousCalendarHandler(service *usecase.AuspiciousService) *AuspiciousCalendarHandler {
	return &AuspiciousCalendarHandler{clock: time.Now, service: service}
}

func (h *AuspiciousCalendarHandler) Activities(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auspicious rule service is unavailable"})
		return
	}
	activities, err := h.service.ListActivities(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดรายการฤกษ์ได้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"activities": activities})
}

func (h *AuspiciousCalendarHandler) Search(c *gin.Context) {
	if h.service == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "auspicious rule service is unavailable"})
		return
	}
	var req domain.AuspiciousSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลค้นหาฤกษ์ไม่ถูกต้อง"})
		return
	}
	response, err := h.service.Search(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

type TaksaRequest struct {
	BirthDate  string `json:"birth_date" form:"birth_date"`
	BirthTime  string `json:"birth_time" form:"birth_time"`
	TargetDate string `json:"target_date" form:"target_date"`
}

func (h *AuspiciousCalendarHandler) Taksa(c *gin.Context) {
	var req TaksaRequest
	if c.Request.Method == http.MethodPost {
		_ = c.ShouldBindJSON(&req)
	} else {
		_ = c.ShouldBindQuery(&req)
	}
	if (req.BirthDate == "") != (req.BirthTime == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "กรุณาระบุวันเกิดและเวลาเกิดเพื่อคำนวณทักษาจร (ใช้เวลา 06:00 น. เป็นเวลาข้ามวัน)"})
		return
	}

	bangkok, _ := time.LoadLocation("Asia/Bangkok")
	// target_date remains accepted for backward compatibility, but Taksa age is
	// always based on the current date in Thailand—not the selected search range.
	currentTime := h.clock().In(bangkok)
	if h.service != nil {
		result, err := h.service.CalculateTaksa(c.Request.Context(), req.BirthDate, req.BirthTime, currentTime)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดข้อความทักษาได้"})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}
	result := usecase.CalculateTaksa(req.BirthDate, req.BirthTime, currentTime)
	c.JSON(http.StatusOK, result)
}

type calendarDayResponse struct {
	Date              string   `json:"date"`
	Day               int      `json:"day"`
	Weekday           string   `json:"weekday"`
	LunarLabel        string   `json:"lunar_label"`
	LunarPhase        string   `json:"lunar_phase"`
	LunarDay          int      `json:"lunar_day"`
	LunarMonth        int      `json:"lunar_month"`
	LunarMonthLabel   string   `json:"lunar_month_label"`
	SecondEighthMonth bool     `json:"second_eighth_month"`
	IsWanPhra         bool     `json:"is_wan_phra"`
	Observances       []string `json:"observances"`
}

// Month returns every day in a Gregorian month paired with its Thai lunar date.
// Dates are interpreted in Asia/Bangkok so the public calendar always follows
// Thai civil-day boundaries.
func (h *AuspiciousCalendarHandler) Month(c *gin.Context) {
	bangkok, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to load Thai timezone"})
		return
	}
	now := h.clock().In(bangkok)
	year, month := now.Year(), int(now.Month())

	if raw := c.Query("year"); raw != "" {
		year, err = strconv.Atoi(raw)
		if err != nil || year < 638 || year > 9999 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "year must be between 638 and 9999"})
			return
		}
	}
	if raw := c.Query("month"); raw != "" {
		month, err = strconv.Atoi(raw)
		if err != nil || month < 1 || month > 12 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "month must be between 1 and 12"})
			return
		}
	}

	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, bangkok)
	days := make([]calendarDayResponse, 0, first.AddDate(0, 1, -1).Day())
	for date := first; date.Month() == first.Month(); date = date.AddDate(0, 0, 1) {
		lunar, conversionErr := thailunar.FromGregorian(date.Year(), int(date.Month()), date.Day())
		if conversionErr != nil {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "unable to convert date to Thai lunar calendar"})
			return
		}
		observances := make([]string, 0, 1)
		if lunar.IsWanPhra() {
			observances = append(observances, "วันพระ")
		}
		days = append(days, calendarDayResponse{
			Date:              date.Format("2006-01-02"),
			Day:               date.Day(),
			Weekday:           lunar.WeekdayName(),
			LunarLabel:        lunar.Phase().String() + " " + strconv.Itoa(lunar.PhaseDay()) + " ค่ำ เดือน " + lunar.MonthName(),
			LunarPhase:        lunar.Phase().String(),
			LunarDay:          lunar.PhaseDay(),
			LunarMonth:        lunar.Month(),
			LunarMonthLabel:   lunar.MonthName(),
			SecondEighthMonth: lunar.IsSecondEighthMonth(),
			IsWanPhra:         lunar.IsWanPhra(),
			Observances:       observances,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year":          year,
		"buddhist_year": year + 543,
		"month":         month,
		"month_name":    thaiGregorianMonthNames[month-1],
		"days":          days,
	})
}

var thaiGregorianMonthNames = [...]string{
	"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
	"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
}
