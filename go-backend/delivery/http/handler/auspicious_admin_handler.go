package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

//go:embed templates/auspicious_admin.html
var auspiciousAdminTemplateFS embed.FS

type AuspiciousAdminHandler struct {
	repo     domain.AuspiciousContentRepository
	auth     *ArticleAdminHandler
	template *template.Template
}

func NewAuspiciousAdminHandler(repo domain.AuspiciousContentRepository, auth *ArticleAdminHandler) (*AuspiciousAdminHandler, error) {
	if repo == nil || auth == nil {
		return nil, nil
	}
	tmpl, err := template.ParseFS(auspiciousAdminTemplateFS, "templates/auspicious_admin.html")
	if err != nil {
		return nil, fmt.Errorf("parse auspicious admin template: %w", err)
	}
	return &AuspiciousAdminHandler{repo: repo, auth: auth, template: tmpl}, nil
}

func (h *AuspiciousAdminHandler) RequireAuth() gin.HandlerFunc { return h.auth.RequireAuth() }

func (h *AuspiciousAdminHandler) Index(c *gin.Context) {
	h.render(c, "")
}

func (h *AuspiciousAdminHandler) Save(c *gin.Context) {
	kind := strings.TrimSpace(c.PostForm("kind"))
	var err error
	switch kind {
	case "text":
		key := strings.TrimSpace(c.PostForm("key"))
		value := strings.TrimSpace(c.PostForm("value"))
		if key == "" || value == "" || utf8.RuneCountInString(value) > 4000 {
			err = fmt.Errorf("ข้อความต้องไม่ว่างและยาวไม่เกิน 4,000 ตัวอักษร")
		} else {
			err = h.repo.UpdateText(c.Request.Context(), key, value)
		}
	case "rule_set":
		id, parseErr := strconv.ParseInt(c.PostForm("id"), 10, 64)
		name := strings.TrimSpace(c.PostForm("name_th"))
		if parseErr != nil || name == "" || utf8.RuneCountInString(name) > 160 {
			err = fmt.Errorf("ชื่อชุดกฎไม่ถูกต้อง")
		} else {
			err = h.repo.UpdateRuleSetText(c.Request.Context(), id, name)
		}
	case "activity":
		code := strings.TrimSpace(c.PostForm("code"))
		name := strings.TrimSpace(c.PostForm("name_th"))
		icon := strings.TrimSpace(c.PostForm("icon"))
		description := strings.TrimSpace(c.PostForm("description_th"))
		scope := strings.TrimSpace(c.PostForm("scope_note_th"))
		safety := strings.TrimSpace(c.PostForm("safety_note_th"))
		if code == "" || name == "" || icon == "" || utf8.RuneCountInString(name) > 120 || utf8.RuneCountInString(icon) > 16 {
			err = fmt.Errorf("ข้อมูลประเภทกิจกรรมไม่ถูกต้อง")
		} else {
			err = h.repo.UpdateActivityText(c.Request.Context(), code, name, icon, description, scope, safety)
		}
	case "rule":
		id, parseErr := strconv.ParseInt(c.PostForm("id"), 10, 64)
		name := strings.TrimSpace(c.PostForm("name_th"))
		description := strings.TrimSpace(c.PostForm("description_th"))
		meaning := strings.TrimSpace(c.PostForm("meaning_th"))
		if parseErr != nil || name == "" || description == "" || meaning == "" || utf8.RuneCountInString(name) > 160 {
			err = fmt.Errorf("ชื่อ คำอธิบาย และความหมายของกฎต้องไม่ว่าง")
		} else {
			err = h.repo.UpdateRuleText(c.Request.Context(), id, name, description, meaning)
		}
	default:
		err = fmt.Errorf("ไม่รู้จักประเภทข้อมูลที่ต้องการบันทึก")
	}
	if err != nil {
		h.render(c, "บันทึกไม่สำเร็จ: "+err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/auspicious-content?success=1")
}

func (h *AuspiciousAdminHandler) render(c *gin.Context, formError string) {
	ctx := c.Request.Context()
	ruleSets, err := h.repo.ListAdminRuleSets(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดชุดกฎฤกษ์ได้: "+err.Error())
		return
	}
	activities, err := h.repo.ListAdminActivities(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดประเภทกิจกรรมได้: "+err.Error())
		return
	}
	rules, err := h.repo.ListAdminRules(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดกฎฤกษ์ได้: "+err.Error())
		return
	}
	texts, err := h.repo.ListTexts(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดข้อความทักษาได้: "+err.Error())
		return
	}
	renderAdminPage(c, h.template, "auspicious_admin.html", auspiciousAdminPage{
		RuleSets: ruleSets, Activities: activities, Rules: rules, Texts: texts,
		Success: c.Query("success") == "1", Error: formError,
	})
}

type auspiciousAdminPage struct {
	RuleSets   []domain.AuspiciousAdminRuleSet
	Activities []domain.AuspiciousAdminActivity
	Rules      []domain.AuspiciousAdminRule
	Texts      []domain.AuspiciousEditableText
	Success    bool
	Error      string
}
