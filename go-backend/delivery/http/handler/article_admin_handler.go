package handler

import (
	"crypto/subtle"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"go-backend/domain"
	"go-backend/repository"
)

//go:embed templates/article_admin.html
var articleAdminTemplateFS embed.FS

// ArticleAdminHandler is only registered when both environment credentials are
// configured. This keeps the management interface unavailable by default.
type ArticleAdminHandler struct {
	repo     repository.ArticleRepository
	username string
	password string
	policy   *bluemonday.Policy
	template *template.Template
}

func NewArticleAdminHandler(repo repository.ArticleRepository, username, password string) (*ArticleAdminHandler, error) {
	if username == "" || password == "" {
		return nil, nil
	}
	templates, err := template.ParseFS(articleAdminTemplateFS, "templates/article_admin.html")
	if err != nil {
		return nil, fmt.Errorf("parse article admin template: %w", err)
	}
	return &ArticleAdminHandler{repo: repo, username: username, password: password, policy: articleSanitizationPolicy(), template: templates}, nil
}

func (h *ArticleAdminHandler) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(h.username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(h.password)) == 1
		if !ok || !usernameMatch || !passwordMatch {
			c.Header("WWW-Authenticate", `Basic realm="Article administration", charset="UTF-8"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

func (h *ArticleAdminHandler) Index(c *gin.Context) {
	articles, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทความได้")
		return
	}
	h.render(c, articleAdminPage{Articles: articles})
}

func (h *ArticleAdminHandler) New(c *gin.Context) {
	h.render(c, articleAdminPage{Article: domain.Article{PublishedAt: time.Now().UTC(), IsPublished: true}})
}

func (h *ArticleAdminHandler) Edit(c *gin.Context) {
	article, ok := h.loadArticle(c)
	if !ok {
		return
	}
	h.render(c, articleAdminPage{Article: *article})
}

func (h *ArticleAdminHandler) Create(c *gin.Context) {
	article, ok := h.articleFromForm(c)
	if !ok {
		return
	}
	if err := h.repo.Create(c.Request.Context(), &article); err != nil {
		h.render(c, articleAdminPage{Article: article, Error: "บันทึกบทความไม่สำเร็จ: " + err.Error()})
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/articles")
}

func (h *ArticleAdminHandler) Update(c *gin.Context) {
	article, ok := h.articleFromForm(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.String(http.StatusBadRequest, "รหัสบทความไม่ถูกต้อง")
		return
	}
	article.ID = id
	if err := h.repo.Update(c.Request.Context(), &article); err != nil {
		h.render(c, articleAdminPage{Article: article, Error: "บันทึกบทความไม่สำเร็จ: " + err.Error()})
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/articles")
}

func (h *ArticleAdminHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.String(http.StatusBadRequest, "รหัสบทความไม่ถูกต้อง")
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.String(http.StatusInternalServerError, "ลบบทความไม่สำเร็จ")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/articles")
}

func (h *ArticleAdminHandler) loadArticle(c *gin.Context) (*domain.Article, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.String(http.StatusBadRequest, "รหัสบทความไม่ถูกต้อง")
		return nil, false
	}
	article, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทความได้")
		return nil, false
	}
	if article == nil {
		c.String(http.StatusNotFound, "ไม่พบบทความ")
		return nil, false
	}
	return article, true
}

func (h *ArticleAdminHandler) articleFromForm(c *gin.Context) (domain.Article, bool) {
	article := domain.Article{
		Slug:        strings.TrimSpace(c.PostForm("slug")),
		Title:       strings.TrimSpace(c.PostForm("title")),
		Excerpt:     strings.TrimSpace(c.PostForm("excerpt")),
		Category:    strings.TrimSpace(c.PostForm("category")),
		ImageURL:    strings.TrimSpace(c.PostForm("image_url")),
		TitleShort:  strings.TrimSpace(c.PostForm("title_short")),
		IsPublished: c.PostForm("is_published") == "on",
		ContentHTML: h.policy.Sanitize(c.PostForm("content_html")),
	}
	if article.Slug == "" || article.Title == "" || article.Excerpt == "" {
		c.String(http.StatusBadRequest, "กรุณากรอก slug, ชื่อบทความ และคำเกริ่นนำ")
		return domain.Article{}, false
	}
	publishedAt, err := time.Parse("2006-01-02", c.PostForm("published_at"))
	if err != nil {
		c.String(http.StatusBadRequest, "วันที่เผยแพร่ไม่ถูกต้อง")
		return domain.Article{}, false
	}
	article.PublishedAt = publishedAt.UTC()
	pinOrder, err := strconv.Atoi(c.PostForm("pin_order"))
	if err != nil || pinOrder < 0 {
		c.String(http.StatusBadRequest, "ลำดับปักหมุดไม่ถูกต้อง")
		return domain.Article{}, false
	}
	article.PinOrder = pinOrder
	return article, true
}

func (h *ArticleAdminHandler) render(c *gin.Context, page articleAdminPage) {
	renderAdminPage(c, h.template, "article_admin.html", page)
}

type articleAdminPage struct {
	Articles []domain.Article
	Article  domain.Article
	Error    string
}
