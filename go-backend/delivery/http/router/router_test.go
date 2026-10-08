package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

func TestAPIContractHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(APIContractHeaders())
	r.GET("/contract", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/contract", nil))
	if recorder.Header().Get("X-API-Version") != domain.APIVersion {
		t.Fatalf("unexpected API version header: %q", recorder.Header().Get("X-API-Version"))
	}
	if recorder.Header().Get("X-Ranking-Contract-Version") != domain.RankingContractVersion {
		t.Fatalf("unexpected ranking contract header: %q", recorder.Header().Get("X-Ranking-Contract-Version"))
	}
}

func TestCanonicalHostRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(CanonicalHostRedirect())
	r.GET("/articles", func(c *gin.Context) { c.Status(http.StatusOK) })

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://www.xn--b3cu8e7ah6h.com/articles?q=seo", nil)
	r.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want 301", recorder.Code)
	}
	if got := recorder.Header().Get("Location"); got != "https://xn--b3cu8e7ah6h.com/articles?q=seo" {
		t.Fatalf("unexpected redirect: %q", got)
	}
}

func TestHTMLPageRedirects(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	redirectPermanent := func(from, to string) {
		r.GET(from, func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, to) })
	}
	redirectPermanent("/about.html", "/about")
	redirectPermanent("/articles.html", "/articles")
	redirectPermanent("/article", "/articles")
	redirectPermanent("/auspicious-calendar.html", "/auspicious-calendar")
	redirectPermanent("/contact.html", "/contact")
	redirectPermanent("/download.html", "/download")
	redirectPermanent("/isan-auspicious-method.html", "/isan-auspicious-method")
	redirectPermanent("/methodology.html", "/methodology")
	redirectPermanent("/privacy.html", "/privacy")
	redirectPermanent("/terms.html", "/terms")

	cases := []struct {
		from string
		to   string
	}{
		{"/about.html", "/about"},
		{"/articles.html", "/articles"},
		{"/article", "/articles"},
		{"/auspicious-calendar.html", "/auspicious-calendar"},
		{"/contact.html", "/contact"},
		{"/download.html", "/download"},
		{"/isan-auspicious-method.html", "/isan-auspicious-method"},
		{"/methodology.html", "/methodology"},
		{"/privacy.html", "/privacy"},
		{"/terms.html", "/terms"},
	}

	for _, tc := range cases {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tc.from, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusMovedPermanently {
			t.Errorf("GET %s status = %d, want 301", tc.from, rec.Code)
		}
		if got := rec.Header().Get("Location"); got != tc.to {
			t.Errorf("GET %s Location = %q, want %q", tc.from, got, tc.to)
		}
	}
}

func TestSearchCrawlerHeadersProtectPrivateRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SearchCrawlerHeaders())
	r.GET("/admin", func(c *gin.Context) { c.Status(http.StatusUnauthorized) })

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/admin", nil))
	if got := recorder.Header().Get("X-Robots-Tag"); !strings.Contains(got, "noindex") {
		t.Fatalf("X-Robots-Tag = %q, want noindex", got)
	}
}

func TestPublicAssetCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(PublicAssetCacheHeaders())
	r.GET("/images/articles/cover.webp", func(c *gin.Context) { c.Status(http.StatusOK) })

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/images/articles/cover.webp", nil))
	if got := recorder.Header().Get("Cache-Control"); !strings.Contains(got, "max-age=604800") {
		t.Fatalf("Cache-Control = %q", got)
	}
}
