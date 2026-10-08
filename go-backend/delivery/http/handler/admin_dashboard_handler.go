package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
	"go-backend/repository"
)

//go:embed templates/admin_dashboard.html
var adminDashboardTemplateFS embed.FS

type AdminDashboardHandler struct {
	repo          repository.ArticleRepository
	nameAdminRepo repository.NameAdminRepository
	auth          *ArticleAdminHandler
	template      *template.Template
}

func NewAdminDashboardHandler(repo repository.ArticleRepository, auth *ArticleAdminHandler) (*AdminDashboardHandler, error) {
	if auth == nil {
		return nil, nil
	}
	templates, err := template.ParseFS(adminDashboardTemplateFS, "templates/admin_dashboard.html")
	if err != nil {
		return nil, fmt.Errorf("parse admin dashboard template: %w", err)
	}
	return &AdminDashboardHandler{repo: repo, auth: auth, template: templates}, nil
}

func (h *AdminDashboardHandler) SetNameAdminRepo(repo repository.NameAdminRepository) {
	h.nameAdminRepo = repo
}

func (h *AdminDashboardHandler) RequireAuth() gin.HandlerFunc { return h.auth.RequireAuth() }

func (h *AdminDashboardHandler) Index(c *gin.Context) {
	stats, err := h.repo.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดข้อมูล dashboard ได้")
		return
	}
	articles, err := h.repo.ListAll(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทความล่าสุดได้")
		return
	}
	if len(articles) > 5 {
		articles = articles[:5]
	}
	media, err := listArticleMedia()
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดคลังรูปได้")
		return
	}

	totalNames := 0
	if h.nameAdminRepo != nil {
		totalNames, _ = h.nameAdminRepo.GetTotalNamesCount(c.Request.Context())
	}

	renderAdminPage(c, h.template, "admin_dashboard.html", adminDashboardPage{
		Stats: stats, MediaCount: len(media), RecentArticles: articles, TotalNames: totalNames,
	})
}

type adminDashboardPage struct {
	Stats          domain.ArticleDashboardStats
	MediaCount     int
	RecentArticles []domain.Article
	TotalNames     int
}
