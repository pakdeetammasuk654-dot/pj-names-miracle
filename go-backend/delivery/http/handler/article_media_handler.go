package handler

import (
	"bytes"
	"crypto/rand"
	"embed"
	"fmt"
	"html/template"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-backend/repository"
)

const (
	articleMediaDir     = "./images/articles"
	articleMediaURLPath = "/images/articles/"
	maxArticleImageSize = 8 << 20
)

//go:embed templates/article_media.html
var articleMediaTemplateFS embed.FS

type ArticleMediaHandler struct {
	repo     repository.ArticleRepository
	auth     *ArticleAdminHandler
	template interface {
		ExecuteTemplate(io.Writer, string, any) error
	}
}

func NewArticleMediaHandler(repo repository.ArticleRepository, auth *ArticleAdminHandler) (*ArticleMediaHandler, error) {
	if auth == nil {
		return nil, nil
	}
	templates, err := template.ParseFS(articleMediaTemplateFS, "templates/article_media.html")
	if err != nil {
		return nil, fmt.Errorf("parse article media template: %w", err)
	}
	return &ArticleMediaHandler{repo: repo, auth: auth, template: templates}, nil
}

func (h *ArticleMediaHandler) RequireAuth() gin.HandlerFunc { return h.auth.RequireAuth() }

func (h *ArticleMediaHandler) Index(c *gin.Context) { h.render(c, "") }

func (h *ArticleMediaHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxArticleImageSize+1024)
	file, _, err := c.Request.FormFile("image")
	if err != nil {
		h.render(c, "เลือกไฟล์รูป JPG, PNG หรือ GIF ขนาดไม่เกิน 8 MB")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxArticleImageSize+1))
	if err != nil || len(data) > maxArticleImageSize {
		h.render(c, "ไฟล์รูปมีขนาดเกิน 8 MB")
		return
	}
	mimeType := http.DetectContentType(data)
	extensions := map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/gif": ".gif"}
	extension, supported := extensions[mimeType]
	if !supported {
		h.render(c, "รองรับเฉพาะไฟล์ JPG, PNG และ GIF")
		return
	}
	if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
		h.render(c, "ไฟล์รูปไม่ถูกต้อง")
		return
	}
	if err := os.MkdirAll(articleMediaDir, 0755); err != nil {
		h.render(c, "ไม่สามารถสร้างพื้นที่เก็บรูปได้")
		return
	}
	filename, err := newMediaFilename(extension)
	if err != nil {
		h.render(c, "ไม่สามารถสร้างชื่อไฟล์ได้")
		return
	}
	if err := os.WriteFile(filepath.Join(articleMediaDir, filename), data, 0644); err != nil {
		h.render(c, "อัปโหลดรูปไม่สำเร็จ")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/articles/media?uploaded=1")
}

func (h *ArticleMediaHandler) Delete(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	if filename == "." || filename != c.Param("filename") || !isSupportedImage(filename) {
		c.String(http.StatusBadRequest, "ชื่อไฟล์ไม่ถูกต้อง")
		return
	}
	imageURL := articleMediaURLPath + filename
	referenced, err := h.repo.IsImageReferenced(c.Request.Context(), imageURL)
	if err != nil {
		c.String(http.StatusInternalServerError, "ตรวจสอบการใช้งานรูปไม่สำเร็จ")
		return
	}
	if referenced {
		h.render(c, "ลบรูปนี้ไม่ได้ เพราะยังถูกใช้งานในบทความ")
		return
	}
	if err := os.Remove(filepath.Join(articleMediaDir, filename)); err != nil {
		if os.IsNotExist(err) {
			c.String(http.StatusNotFound, "ไม่พบรูป")
			return
		}
		c.String(http.StatusInternalServerError, "ลบรูปไม่สำเร็จ")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/articles/media")
}

func (h *ArticleMediaHandler) render(c *gin.Context, message string) {
	items, err := listArticleMedia()
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดคลังรูปได้")
		return
	}
	renderAdminPage(c, h.template, "article_media.html", articleMediaPage{Items: items, Message: message, Uploaded: c.Query("uploaded") == "1"})
}

type articleMediaPage struct {
	Items    []articleMediaItem
	Message  string
	Uploaded bool
}

type articleMediaItem struct {
	Name     string
	URL      string
	Modified time.Time
}

func listArticleMedia() ([]articleMediaItem, error) {
	entries, err := os.ReadDir(articleMediaDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []articleMediaItem{}, nil
		}
		return nil, err
	}
	items := make([]articleMediaItem, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !isSupportedImage(entry.Name()) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		items = append(items, articleMediaItem{Name: entry.Name(), URL: articleMediaURLPath + entry.Name(), Modified: info.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Modified.After(items[j].Modified) })
	return items, nil
}

func isSupportedImage(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func newMediaFilename(extension string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x%s", bytes, extension), nil
}
