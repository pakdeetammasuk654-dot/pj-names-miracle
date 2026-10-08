package handler

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
	"go-backend/repository"
	"go-backend/usecase"
)

const maxSupportImageBytes = 8 << 20

//go:embed templates/support_chat_admin.html templates/support_chat_customer.html
var supportChatTemplateFS embed.FS

type SupportChatHandler struct {
	repo     repository.SupportChatRepository
	auth     *ArticleAdminHandler
	template *template.Template
	push     usecase.PushNotificationService
}

func customerAvatarURL(avatar string) string {
	switch strings.ToLower(strings.TrimSpace(avatar)) {
	case "cat":
		return "/images/chat_avatar_cat.png"
	case "dog":
		return "/images/chat_avatar_dog.png"
	case "rabbit":
		return "/images/chat_avatar_rabbit.png"
	case "bear":
		return "/images/chat_avatar_bear.png"
	case "panda":
		return "/images/chat_avatar_panda.png"
	case "male":
		return "/images/chat_avatar_dog.png"
	case "female":
		return "/images/chat_avatar_cat.png"
	default:
		return "/images/chat_avatar_cat.png"
	}
}

func normalizeAvatar(avatar string) string {
	clean := strings.TrimSpace(strings.ToLower(avatar))
	switch {
	case strings.Contains(clean, "cat"):
		return "cat"
	case strings.Contains(clean, "dog"):
		return "dog"
	case strings.Contains(clean, "rabbit"):
		return "rabbit"
	case strings.Contains(clean, "bear"):
		return "bear"
	case strings.Contains(clean, "panda"):
		return "panda"
	case strings.Contains(clean, "male") && !strings.Contains(clean, "female"):
		return "dog"
	case strings.Contains(clean, "female"):
		return "cat"
	default:
		return "cat"
	}
}

func NewSupportChatHandler(repo repository.SupportChatRepository, auth *ArticleAdminHandler, push ...usecase.PushNotificationService) (*SupportChatHandler, error) {
	tmpl := template.New("support_chat").Funcs(template.FuncMap{
		"customerAvatarURL": customerAvatarURL,
	})
	templates, err := tmpl.ParseFS(supportChatTemplateFS, "templates/support_chat_admin.html", "templates/support_chat_customer.html")
	if err != nil {
		return nil, fmt.Errorf("parse support chat templates: %w", err)
	}
	var pushSvc usecase.PushNotificationService
	if len(push) > 0 {
		pushSvc = push[0]
	}
	return &SupportChatHandler{repo: repo, auth: auth, template: templates, push: pushSvc}, nil
}

func (h *SupportChatHandler) RequireAuth() gin.HandlerFunc {
	if h.auth == nil {
		return func(c *gin.Context) { c.AbortWithStatus(http.StatusNotFound) }
	}
	return h.auth.RequireAuth()
}

// CustomerChatPage is the public web support entry point. It persists only an
// opaque conversation token in the visitor's browser; no account is required.
func (h *SupportChatHandler) CustomerChatPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.template.ExecuteTemplate(c.Writer, "support_chat_customer.html", nil); err != nil {
		c.Error(err)
	}
}

func (h *SupportChatHandler) AdminIndex(c *gin.Context) {
	inbox := c.DefaultQuery("view", "needs_reply")
	if !validInboxView(inbox) {
		c.String(http.StatusBadRequest, "ตัวกรองไม่ถูกต้อง")
		return
	}
	items, err := h.repo.ListConversations(c.Request.Context(), inbox)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทสนทนาได้")
		return
	}
	stats, err := h.repo.GetInboxStats(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดจำนวนบทสนทนาได้")
		return
	}
	h.render(c, "support_chat_admin.html", supportChatAdminPage{Conversations: items, Inbox: inbox, Stats: stats})
}

func (h *SupportChatHandler) AdminConversation(c *gin.Context) {
	conversation, err := h.repo.GetConversation(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทสนทนาได้")
		return
	}
	if conversation == nil {
		c.String(http.StatusNotFound, "ไม่พบบทสนทนา")
		return
	}
	inbox := c.DefaultQuery("view", "needs_reply")
	if !validInboxView(inbox) {
		c.String(http.StatusBadRequest, "ตัวกรองไม่ถูกต้อง")
		return
	}
	conversations, err := h.repo.ListConversations(c.Request.Context(), inbox)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดรายการบทสนทนาได้")
		return
	}
	_ = h.repo.MarkAdminRead(c.Request.Context(), conversation.ID)
	messages, err := h.repo.ListMessages(c.Request.Context(), conversation.ID)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดข้อความได้")
		return
	}
	stats, err := h.repo.GetInboxStats(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดจำนวนบทสนทนาได้")
		return
	}
	h.render(c, "support_chat_admin.html", supportChatAdminPage{Conversations: conversations, Active: conversation, Messages: messages, Inbox: inbox, Stats: stats})
}

func (h *SupportChatHandler) AdminInboxStats(c *gin.Context) {
	stats, err := h.repo.GetInboxStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดจำนวนบทสนทนาได้"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

func (h *SupportChatHandler) AdminMessagesJSON(c *gin.Context) {
	conversation, err := h.repo.GetConversation(c.Request.Context(), c.Param("id"))
	if err != nil || conversation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบบทสนทนา"})
		return
	}
	_ = h.repo.MarkAdminRead(c.Request.Context(), conversation.ID)
	messages, err := h.repo.ListMessages(c.Request.Context(), conversation.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดข้อความได้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":              conversation.ID,
		"customer_name":   conversation.CustomerName,
		"customer_avatar": conversation.CustomerAvatar,
		"status":          conversation.Status,
		"messages":        messages,
	})
}

func (h *SupportChatHandler) AdminSend(c *gin.Context) {
	conversation, err := h.repo.GetConversation(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถตรวจสอบบทสนทนาได้")
		return
	}
	if conversation == nil {
		c.String(http.StatusNotFound, "ไม่พบบทสนทนา")
		return
	}
	imageURL, err := saveSupportImage(c)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.createMessage(c, c.Param("id"), "admin", c.PostForm("body"), imageURL); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/chat/"+c.Param("id")+"?view=active")
}

func (h *SupportChatHandler) AdminClose(c *gin.Context) {
	if err := h.repo.SetConversationStatus(c.Request.Context(), c.Param("id"), "closed"); err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "ไม่พบบทสนทนา")
			return
		}
		c.String(http.StatusInternalServerError, "ไม่สามารถปิดเคสได้")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/chat?view=closed")
}

func (h *SupportChatHandler) AdminDeleteMessage(c *gin.Context) {
	if err := h.deleteAdminMessage(c); err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "ไม่พบข้อความ")
			return
		}
		c.String(http.StatusInternalServerError, "ไม่สามารถลบข้อความได้")
		return
	}
	c.Redirect(http.StatusSeeOther, "/admin/chat/"+c.Param("id")+"?view="+c.DefaultQuery("view", "active"))
}

func (h *SupportChatHandler) AdminConversationsJSON(c *gin.Context) {
	inbox := c.DefaultQuery("view", "needs_reply")
	if !validInboxView(inbox) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ตัวกรองไม่ถูกต้อง"})
		return
	}
	items, err := h.repo.ListConversations(c.Request.Context(), inbox)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดบทสนทนาได้"})
		return
	}
	stats, err := h.repo.GetInboxStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดจำนวนบทสนทนาได้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"inbox":         inbox,
		"conversations": items,
		"stats":         stats,
	})
}

func (h *SupportChatHandler) AdminSendJSON(c *gin.Context) {
	conversation, err := h.repo.GetConversation(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบบทสนทนาได้"})
		return
	}
	if conversation == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบบทสนทนา"})
		return
	}
	imageURL, err := saveSupportImage(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.createMessage(c, c.Param("id"), "admin", c.PostForm("body"), imageURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (h *SupportChatHandler) AdminCloseJSON(c *gin.Context) {
	if err := h.repo.SetConversationStatus(c.Request.Context(), c.Param("id"), "closed"); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบบทสนทนา"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถปิดเคสได้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *SupportChatHandler) AdminDeleteMessageJSON(c *gin.Context) {
	if err := h.deleteAdminMessage(c); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "ไม่พบข้อความ"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อความได้"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *SupportChatHandler) deleteAdminMessage(c *gin.Context) error {
	imageURL, err := h.repo.DeleteMessage(c.Request.Context(), c.Param("id"), c.Param("messageID"))
	if err != nil {
		return err
	}
	// The database deletion is authoritative for every client. A failed cleanup
	// only leaves an inaccessible upload behind; it must not report a deleted
	// message as still present to the administrator.
	_ = removeSupportImage(imageURL)
	return nil
}

type adminDeviceTokenRequest struct {
	Platform    string `json:"platform"`
	DeviceToken string `json:"device_token"`
	AppVersion  string `json:"app_version"`
}

func (h *SupportChatHandler) AdminRegisterDeviceToken(c *gin.Context) {
	var req adminDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูลไม่ถูกต้อง"})
		return
	}
	token := strings.TrimSpace(req.DeviceToken)
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_token ว่างเปล่า"})
		return
	}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	if platform == "" {
		platform = "ios_admin"
	}
	if err := h.repo.RegisterAdminDeviceToken(c.Request.Context(), platform, token, req.AppVersion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถบันทึก device token ได้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *SupportChatHandler) AdminAuthLogin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"role":     "admin",
		"username": "admin",
		"name":     "อ.ทญา ว่องไว",
	})
}

func (h *SupportChatHandler) CreateCustomerConversation(c *gin.Context) {
	name := strings.TrimSpace(c.PostForm("customer_name"))
	if len([]rune(name)) > 80 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ชื่อที่ใช้เรียกยาวเกินไป"})
		return
	}
	avatar := normalizeAvatar(c.PostForm("customer_avatar"))
	conversation := domain.SupportConversation{
		ID:             newOpaqueID(),
		CustomerToken:  newOpaqueID(),
		CustomerName:   name,
		CustomerAvatar: avatar,
	}
	if err := h.repo.CreateConversation(c.Request.Context(), conversation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเริ่มบทสนทนาได้"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"conversation_id": conversation.ID,
		"customer_token":  conversation.CustomerToken,
		"customer_avatar": conversation.CustomerAvatar,
	})
}

func (h *SupportChatHandler) CustomerMessages(c *gin.Context) {
	conversation := h.customerConversation(c)
	if conversation == nil {
		return
	}
	_ = h.repo.MarkCustomerRead(c.Request.Context(), conversation.ID)
	messages, err := h.repo.ListMessages(c.Request.Context(), conversation.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดข้อความได้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"messages":              messages,
		"customer_avatar":       conversation.CustomerAvatar,
		"customer_name":         conversation.CustomerName,
		"customer_unread_count": 0,
		"status":                conversation.Status,
	})
}

func (h *SupportChatHandler) CustomerUnread(c *gin.Context) {
	conversation := h.customerConversation(c)
	if conversation == nil {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"unread_count":    conversation.CustomerUnreadCount,
		"status":          conversation.Status,
		"customer_avatar": conversation.CustomerAvatar,
	})
}

func (h *SupportChatHandler) DeleteCustomerConversation(c *gin.Context) {
	conversation := h.customerConversation(c)
	if conversation == nil {
		return
	}

	imageURLs, err := h.repo.ListCustomerImageURLs(c.Request.Context(), conversation.ID, conversation.CustomerToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถเตรียมลบข้อมูลแชตได้"})
		return
	}
	for _, imageURL := range imageURLs {
		if err := removeSupportImage(imageURL); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบรูปในแชตได้ กรุณาลองใหม่"})
			return
		}
	}
	if err := h.repo.DeleteConversationForCustomer(c.Request.Context(), conversation.ID, conversation.CustomerToken); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "บทสนทนานี้ถูกลบไปแล้ว"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลบข้อมูลแชตได้ กรุณาลองใหม่"})
		return
	}
	c.Status(http.StatusNoContent)
}

func removeSupportImage(imageURL string) error {
	const prefix = "/chat-uploads/"
	if !strings.HasPrefix(imageURL, prefix) {
		return nil
	}
	filename := strings.TrimPrefix(imageURL, prefix)
	if filename == "" || filepath.Base(filename) != filename {
		return fmt.Errorf("invalid support image path")
	}
	err := os.Remove(filepath.Join("./chat_uploads", filename))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (h *SupportChatHandler) CustomerSend(c *gin.Context) {
	conversation := h.customerConversation(c)
	if conversation == nil {
		return
	}
	imageURL, err := saveSupportImage(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.createMessage(c, conversation.ID, "customer", c.PostForm("body"), imageURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *SupportChatHandler) customerConversation(c *gin.Context) *domain.SupportConversation {
	token := strings.TrimSpace(c.GetHeader("X-Chat-Token"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token ของบทสนทนาไม่ถูกต้อง"})
		return nil
	}
	conversation, err := h.repo.GetConversationForCustomer(c.Request.Context(), c.Param("id"), token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบบทสนทนาได้"})
		return nil
	}
	if conversation == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token ของบทสนทนาไม่ถูกต้อง"})
		return nil
	}
	return conversation
}

func (h *SupportChatHandler) RegisterCustomerDeviceToken(c *gin.Context) {
	conversation := h.customerConversation(c)
	if conversation == nil {
		return
	}
	var req domain.RegisterDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ข้อมูล token ไม่ถูกต้อง"})
		return
	}
	req.DeviceToken = strings.TrimSpace(req.DeviceToken)
	if req.DeviceToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_token ว่างเปล่า"})
		return
	}
	platform := strings.ToLower(strings.TrimSpace(req.Platform))
	if platform == "" {
		platform = "android"
	}
	if err := h.repo.RegisterDeviceToken(c.Request.Context(), conversation.ID, platform, req.DeviceToken, req.AppVersion); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถลงทะเบียน device token ได้"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *SupportChatHandler) createMessage(c *gin.Context, conversationID, sender, body, imageURL string) error {
	body = strings.TrimSpace(body)
	if len([]rune(body)) > 4000 {
		return fmt.Errorf("ข้อความยาวเกิน 4,000 ตัวอักษร")
	}
	if body == "" && imageURL == "" {
		return fmt.Errorf("กรุณาพิมพ์ข้อความหรือแนบรูป")
	}
	msg := domain.SupportMessage{
		ID:             newOpaqueID(),
		ConversationID: conversationID,
		Sender:         sender,
		Body:           body,
		ImageURL:       imageURL,
	}
	if err := h.repo.CreateMessage(c.Request.Context(), msg); err != nil {
		return err
	}
	if sender == "admin" && h.push != nil {
		h.push.SendAdminReplyNotification(context.Background(), conversationID, msg)
	} else if sender == "customer" && h.push != nil {
		h.push.SendCustomerMessageNotification(context.Background(), conversationID, msg)
	}
	return nil
}

func (h *SupportChatHandler) render(c *gin.Context, name string, page any) {
	renderAdminPage(c, h.template, name, page)
}

func newOpaqueID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("cannot create secure support chat id")
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16]))
}

func saveSupportImage(c *gin.Context) (string, error) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSupportImageBytes+1024*1024)
	file, header, err := c.Request.FormFile("image")
	if err == http.ErrMissingFile || err == http.ErrNotMultipart {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("ไม่สามารถอ่านรูปได้ (ขนาดสูงสุด 8 MB)")
	}
	defer file.Close()
	if header.Size > maxSupportImageBytes {
		return "", fmt.Errorf("รูปต้องมีขนาดไม่เกิน 8 MB")
	}
	content, err := io.ReadAll(io.LimitReader(file, maxSupportImageBytes+1))
	if err != nil {
		return "", fmt.Errorf("ไม่สามารถอ่านรูปได้")
	}
	if len(content) > maxSupportImageBytes {
		return "", fmt.Errorf("รูปต้องมีขนาดไม่เกิน 8 MB")
	}
	ext := ""
	switch http.DetectContentType(content) {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		return "", fmt.Errorf("รองรับเฉพาะ JPG, PNG และ WEBP")
	}
	if err := os.MkdirAll("./chat_uploads", 0750); err != nil {
		return "", fmt.Errorf("ไม่สามารถจัดเก็บรูปได้")
	}
	filename := newOpaqueID() + ext
	if err := os.WriteFile(filepath.Join("./chat_uploads", filename), content, 0640); err != nil {
		return "", fmt.Errorf("ไม่สามารถจัดเก็บรูปได้")
	}
	return "/chat-uploads/" + filename, nil
}

type supportChatAdminPage struct {
	Conversations []domain.SupportConversation
	Active        *domain.SupportConversation
	Messages      []domain.SupportMessage
	Inbox         string
	Stats         domain.SupportChatInboxStats
}

func validInboxView(view string) bool {
	return view == "needs_reply" || view == "active" || view == "closed"
}
