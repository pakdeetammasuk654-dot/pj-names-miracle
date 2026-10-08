package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
	"go-backend/usecase"
)

//go:embed templates/name_admin.html
var nameAdminTemplateFS embed.FS

type NameAdminHandler struct {
	useCase  usecase.NameAdminUseCase
	auth     *ArticleAdminHandler
	template *template.Template
}

func NewNameAdminHandler(useCase usecase.NameAdminUseCase, auth *ArticleAdminHandler) (*NameAdminHandler, error) {
	tmpl, err := template.ParseFS(nameAdminTemplateFS, "templates/name_admin.html")
	if err != nil {
		return nil, fmt.Errorf("parse name admin template: %w", err)
	}
	return &NameAdminHandler{
		useCase:  useCase,
		auth:     auth,
		template: tmpl,
	}, nil
}

func (h *NameAdminHandler) RequireAuth() gin.HandlerFunc {
	return h.auth.RequireAuth()
}

type nameAdminPageData struct {
	RecentNames    []domain.AdminNameItem
	TotalCount     int
	SearchQuery    string
	SuccessMessage string
	ErrorMessage   string
}

func (h *NameAdminHandler) Index(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	names, total, err := h.useCase.ListNames(c.Request.Context(), q, 1, 40)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดข้อมูลรายชื่อได้: "+err.Error())
		return
	}

	data := nameAdminPageData{
		RecentNames:    names,
		TotalCount:     total,
		SearchQuery:    q,
		SuccessMessage: c.Query("success"),
		ErrorMessage:   c.Query("error"),
	}

	renderAdminPage(c, h.template, "name_admin.html", data)
}

func (h *NameAdminHandler) AnalyzeAPI(c *gin.Context) {
	var req domain.NameAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": "รูปแบบข้อมูลไม่ถูกต้อง: " + err.Error()})
		return
	}

	result, err := h.useCase.AnalyzeName(c.Request.Context(), req.THName, req.AutoMeaning, req.Gender)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": true, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": false, "data": result})
}

func (h *NameAdminHandler) Create(c *gin.Context) {
	thname := strings.TrimSpace(c.PostForm("thname"))
	meaning := strings.TrimSpace(c.PostForm("meaning"))
	gender := strings.TrimSpace(c.PostForm("gender"))

	satSum, _ := strconv.Atoi(c.PostForm("sat_sum"))
	shaSum, _ := strconv.Atoi(c.PostForm("sha_sum"))
	euphony, _ := strconv.Atoi(c.PostForm("euphony_score"))
	ease, _ := strconv.Atoi(c.PostForm("pronunciation_ease"))
	rhythm, _ := strconv.Atoi(c.PostForm("rhythm_score"))

	req := &domain.CreateAdminNameRequest{
		THName:            thname,
		Meaning:           meaning,
		Gender:            gender,
		SatSum:            satSum,
		ShaSum:            shaSum,
		EuphonyScore:      euphony,
		PronunciationEase: ease,
		RhythmScore:       rhythm,
	}

	_, err := h.useCase.SaveName(c.Request.Context(), req)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/admin/names?error="+url.QueryEscape("บันทึกไม่สำเร็จ: "+err.Error()))
		return
	}

	msg := fmt.Sprintf("เพิ่มชื่อมงคล %q (เลขศาสตร์: %d, พลังเงา: %d) เข้าสู่ระบบเรียบร้อยแล้ว", thname, satSum, shaSum)
	c.Redirect(http.StatusSeeOther, "/admin/names?success="+url.QueryEscape(msg))
}
