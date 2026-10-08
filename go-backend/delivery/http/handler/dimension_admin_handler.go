package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
	"go-backend/repository"
)

//go:embed templates/dimension_admin.html
var dimensionAdminTemplateFS embed.FS

type DimensionAdminHandler struct {
	repo     repository.DimensionLabelRepository
	auth     *ArticleAdminHandler
	template *template.Template
}

func NewDimensionAdminHandler(repo repository.DimensionLabelRepository, auth *ArticleAdminHandler) (*DimensionAdminHandler, error) {
	if auth == nil {
		return nil, nil
	}
	tmpl, err := template.ParseFS(dimensionAdminTemplateFS, "templates/dimension_admin.html")
	if err != nil {
		return nil, fmt.Errorf("parse dimension admin template: %w", err)
	}
	return &DimensionAdminHandler{repo: repo, auth: auth, template: tmpl}, nil
}

func (h *DimensionAdminHandler) RequireAuth() gin.HandlerFunc { return h.auth.RequireAuth() }

func (h *DimensionAdminHandler) Index(c *gin.Context) {
	groups, err := h.repo.GetGrouped(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดข้อมูลมิติพลัง 4 ด้านได้: "+err.Error())
		return
	}
	renderAdminPage(c, h.template, "dimension_admin.html", dimensionAdminPage{
		Groups:  groups,
		Success: c.Query("success") == "1",
	})
}

func (h *DimensionAdminHandler) Save(c *gin.Context) {
	items, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดข้อมูลเดิมเพื่ออัปเดตได้")
		return
	}

	for _, item := range items {
		idStr := strconv.Itoa(item.ID)
		label := strings.TrimSpace(c.PostForm("label_" + idStr))
		desc := strings.TrimSpace(c.PostForm("description_" + idStr))
		minScoreStr := c.PostForm("min_score_" + idStr)

		minScore, err := strconv.ParseFloat(minScoreStr, 64)
		if err != nil || label == "" {
			continue
		}

		if err := h.repo.Update(c.Request.Context(), item.ID, minScore, label, desc); err != nil {
			groups, _ := h.repo.GetGrouped(c.Request.Context())
			renderAdminPage(c, h.template, "dimension_admin.html", dimensionAdminPage{
				Groups: groups,
				Error:  "บันทึกไม่สำเร็จ: " + err.Error(),
			})
			return
		}
	}

	c.Redirect(http.StatusFound, "/admin/dimension-labels?success=1")
}

type dimensionAdminPage struct {
	Groups  []domain.DimensionGroup
	Success bool
	Error   string
}
