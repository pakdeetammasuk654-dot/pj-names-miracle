package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

type fakeArticleRepository struct{ articles []domain.Article }

func (r fakeArticleRepository) ListPublished(context.Context) ([]domain.Article, error) {
	return r.articles, nil
}
func (r fakeArticleRepository) GetPublishedBySlug(_ context.Context, slug string) (*domain.Article, error) {
	for _, article := range r.articles {
		if article.Slug == slug {
			return &article, nil
		}
	}
	return nil, nil
}
func (r fakeArticleRepository) ListAll(context.Context) ([]domain.Article, error) {
	return r.articles, nil
}
func (r fakeArticleRepository) GetByID(context.Context, int64) (*domain.Article, error) {
	return nil, nil
}
func (r fakeArticleRepository) Create(context.Context, *domain.Article) error { return nil }
func (r fakeArticleRepository) Update(context.Context, *domain.Article) error { return nil }
func (r fakeArticleRepository) Delete(context.Context, int64) error           { return nil }
func (r fakeArticleRepository) IsImageReferenced(context.Context, string) (bool, error) {
	return false, nil
}
func (r fakeArticleRepository) GetDashboardStats(context.Context) (domain.ArticleDashboardStats, error) {
	return domain.ArticleDashboardStats{}, nil
}

func TestArticleDetailSanitizesContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, err := NewArticleHandler(fakeArticleRepository{articles: []domain.Article{
		{
			ID: 1, Slug: "safe", Title: "บทความทดสอบ", Excerpt: "คำเกริ่น", Category: "เคล็ดลับ", ImageURL: "/images/articles/cover.jpg",
			PublishedAt: time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC), IsPublished: true,
			ContentHTML: `<p>เนื้อหาปลอดภัย</p><script>alert('xss')</script>`,
		},
		{ID: 2, Slug: "related", Title: "บทความที่เกี่ยวข้อง", Category: "เคล็ดลับ", IsPublished: true},
	}})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/articles/:slug", handler.Detail)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/articles/safe", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	if strings.Contains(recorder.Body.String(), "<script>") {
		t.Fatal("unsafe script was rendered")
	}
	if !strings.Contains(recorder.Body.String(), "เนื้อหาปลอดภัย") {
		t.Fatal("safe article content missing")
	}
	if !strings.Contains(recorder.Body.String(), `"@type":"Article"`) {
		t.Fatal("article structured data missing")
	}
	if !strings.Contains(recorder.Body.String(), `"@type":"BreadcrumbList"`) {
		t.Fatal("breadcrumb structured data missing")
	}
	if !strings.Contains(recorder.Body.String(), "ทญา ว่องไว") {
		t.Fatal("visible author attribution missing")
	}
	if !strings.Contains(recorder.Body.String(), "/articles/related") {
		t.Fatal("related article link missing")
	}
	if !strings.Contains(recorder.Body.String(), "แชร์บทความนี้") {
		t.Fatal("share controls missing")
	}
}

func TestArticleSitemapContainsPublishedArticle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, err := NewArticleHandler(fakeArticleRepository{articles: []domain.Article{{Slug: "ชื่อ-ทดสอบ", PublishedAt: time.Now(), IsPublished: true}}})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/sitemap.xml", handler.Sitemap)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "articles/%E0%B8%8A%E0%B8%B7%E0%B9%88%E0%B8%AD-%E0%B8%97%E0%B8%94%E0%B8%AA%E0%B8%AD%E0%B8%9A") {
		t.Fatal("article URL missing from sitemap")
	}
	for _, path := range []string{"/articles", "/download", "/auspicious-calendar", "/isan-auspicious-method", "/methodology", "/about", "/contact", "/privacy", "/terms"} {
		if !strings.Contains(recorder.Body.String(), path) {
			t.Fatalf("static public URL %s missing from sitemap", path)
		}
	}
}

func TestArticleSEOTitleShortensLongHeadline(t *testing.T) {
	article := domain.Article{
		Title:      "ถอดรหัสพลังมหาเศรษฐีและการค้นหาชื่อมงคลคู่บารมีด้วยระบบชื่อดีสี่มิติฉบับสมบูรณ์",
		TitleShort: "อัศวโภคิน",
	}
	got := articleSEOTitle(article)
	if got != "วิเคราะห์ชื่อ อัศวโภคิน ด้วยเลขศาสตร์ | ชื่อดี" {
		t.Fatalf("unexpected SEO title: %q", got)
	}
}

func TestEnhanceArticleImagesAddsUsefulAttributes(t *testing.T) {
	got := enhanceArticleImages(`<p><img src="/images/articles/a.png" alt="a.png"></p>`, "หัวข้อทดสอบ")
	for _, expected := range []string{`loading="lazy"`, `decoding="async"`, `alt="ภาพประกอบบทความ หัวข้อทดสอบ"`} {
		if !strings.Contains(got, expected) {
			t.Fatalf("enhanced image missing %s: %s", expected, got)
		}
	}
}

func TestSelectRelatedArticlesPrefersCategoryAndContentSimilarity(t *testing.T) {
	current := domain.Article{ID: 1, Category: "ตั้งชื่อมงคล", Title: "เนเน่ รอยัล เปลี่ยนชื่อ", Excerpt: "วิเคราะห์ชื่อและความสำเร็จ"}
	articles := []domain.Article{
		current,
		{ID: 2, Category: "ตั้งชื่อมงคล", Title: "แนวทางเลือกชื่อเด็ก", Excerpt: "หลักการตั้งชื่อมงคล"},
		{ID: 3, Category: "คนดัง", Title: "เนเน่ รอยัล กับชื่อใหม่", Excerpt: "ความสำเร็จหลังเปลี่ยนชื่อ"},
		{ID: 4, Category: "กีฬา", Title: "ผลการแข่งขันฟุตบอล", Excerpt: "ข่าวกีฬาวันนี้"},
	}

	related := selectRelatedArticles(current, articles)
	if len(related) != 3 {
		t.Fatalf("related count = %d, want 3", len(related))
	}
	if related[0].ID != 2 {
		t.Fatalf("same-category article should be first, got ID %d", related[0].ID)
	}
	if related[1].ID != 3 {
		t.Fatalf("similar article should be second, got ID %d", related[1].ID)
	}
}
