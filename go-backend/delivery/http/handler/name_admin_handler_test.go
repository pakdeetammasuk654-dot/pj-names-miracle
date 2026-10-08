package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

type mockNameAdminUseCase struct {
	names []domain.AdminNameItem
}

func (m *mockNameAdminUseCase) AnalyzeName(ctx context.Context, thname string, autoMeaning bool, gender string) (*domain.NameAnalysisResult, error) {
	return &domain.NameAnalysisResult{
		THName:   thname,
		SatSum:   16,
		ShaSum:   30,
		SatPairs: []string{"16"},
		ShaPairs: []string{"30"},
		Meaning:  "ความหมายทดสอบ",
		Gender:   gender,
	}, nil
}

func (m *mockNameAdminUseCase) SaveName(ctx context.Context, req *domain.CreateAdminNameRequest) (*domain.AdminNameItem, error) {
	item := domain.AdminNameItem{
		NameID:  123,
		THName:  req.THName,
		Meaning: req.Meaning,
		Gender:  req.Gender,
		SatSum:  req.SatSum,
		ShaSum:  req.ShaSum,
	}
	m.names = append(m.names, item)
	return &item, nil
}

func (m *mockNameAdminUseCase) ListNames(ctx context.Context, search string, page int, pageSize int) ([]domain.AdminNameItem, int, error) {
	return m.names, len(m.names), nil
}

func (m *mockNameAdminUseCase) GetTotalNamesCount(ctx context.Context) (int, error) {
	return len(m.names), nil
}

func TestNameAdminHandler_AnalyzeAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := &mockNameAdminUseCase{}
	h, err := NewNameAdminHandler(mockUC, nil)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	router := gin.New()
	router.POST("/admin/names/analyze", h.AnalyzeAPI)

	payload := domain.NameAnalysisRequest{
		THName:      "เจเน่",
		AutoMeaning: true,
		Gender:      "ค",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/admin/names/analyze", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Error bool                      `json:"error"`
		Data  domain.NameAnalysisResult `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if res.Error {
		t.Errorf("expected error false, got true")
	}
	if res.Data.THName != "เจเน่" {
		t.Errorf("expected THName 'เจเน่', got %s", res.Data.THName)
	}
	if res.Data.SatSum != 16 {
		t.Errorf("expected SatSum 16, got %d", res.Data.SatSum)
	}
}

func TestNameAdminHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockUC := &mockNameAdminUseCase{}
	h, err := NewNameAdminHandler(mockUC, nil)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	router := gin.New()
	router.POST("/admin/names/create", h.Create)

	form := url.Values{
		"thname":   {"เจเน่"},
		"meaning":  {"ความหมายทดสอบ"},
		"gender":   {"ค"},
		"sat_sum":  {"16"},
		"sha_sum":  {"30"},
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/names/create", bytes.NewBufferString(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 redirect, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Errorf("expected Location header redirect")
	}

	if len(mockUC.names) != 1 {
		t.Errorf("expected 1 name saved, got %d", len(mockUC.names))
	}
}
