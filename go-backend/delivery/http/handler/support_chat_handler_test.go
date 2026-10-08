package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go-backend/domain"
)

type supportChatRepoStub struct {
	created          domain.SupportConversation
	owned            bool
	deleted          bool
	deletedMessageID string
}

func (s *supportChatRepoStub) CreateConversation(_ context.Context, c domain.SupportConversation) error {
	s.created = c
	return nil
}
func (s *supportChatRepoStub) GetConversation(context.Context, string) (*domain.SupportConversation, error) {
	return nil, nil
}
func (s *supportChatRepoStub) GetConversationForCustomer(_ context.Context, id, token string) (*domain.SupportConversation, error) {
	if s.owned && id == s.created.ID && token == s.created.CustomerToken {
		return &s.created, nil
	}
	return nil, nil
}
func (s *supportChatRepoStub) ListCustomerImageURLs(_ context.Context, id, token string) ([]string, error) {
	if s.owned && id == s.created.ID && token == s.created.CustomerToken {
		return []string{}, nil
	}
	return nil, nil
}
func (s *supportChatRepoStub) DeleteConversationForCustomer(_ context.Context, id, token string) error {
	if s.owned && id == s.created.ID && token == s.created.CustomerToken {
		s.deleted = true
	}
	return nil
}
func (s *supportChatRepoStub) ListConversations(context.Context, string) ([]domain.SupportConversation, error) {
	return nil, nil
}
func (s *supportChatRepoStub) ListMessages(context.Context, string) ([]domain.SupportMessage, error) {
	return nil, nil
}
func (s *supportChatRepoStub) CreateMessage(context.Context, domain.SupportMessage) error { return nil }
func (s *supportChatRepoStub) DeleteMessage(_ context.Context, _, messageID string) (string, error) {
	s.deletedMessageID = messageID
	return "", nil
}
func (s *supportChatRepoStub) MarkCustomerRead(context.Context, string) error {
	s.created.CustomerUnreadCount = 0
	return nil
}
func (s *supportChatRepoStub) MarkAdminRead(context.Context, string) error {
	s.created.AdminUnreadCount = 0
	return nil
}
func (s *supportChatRepoStub) SetConversationStatus(context.Context, string, string) error {
	return nil
}
func (s *supportChatRepoStub) GetInboxStats(context.Context) (domain.SupportChatInboxStats, error) {
	return domain.SupportChatInboxStats{}, nil
}
func (s *supportChatRepoStub) RegisterDeviceToken(context.Context, string, string, string, string) error {
	return nil
}
func (s *supportChatRepoStub) GetActiveDeviceTokens(context.Context, string) ([]domain.SupportChatDeviceToken, error) {
	return nil, nil
}
func (s *supportChatRepoStub) DeactivateDeviceToken(context.Context, string) error {
	return nil
}
func (s *supportChatRepoStub) RegisterAdminDeviceToken(context.Context, string, string, string) error {
	return nil
}
func (s *supportChatRepoStub) GetActiveAdminDeviceTokens(context.Context) ([]domain.SupportChatDeviceToken, error) {
	return nil, nil
}
func (s *supportChatRepoStub) DeactivateAdminDeviceToken(context.Context, string) error {
	return nil
}

func TestCreateCustomerConversationReturnsOpaqueCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &supportChatRepoStub{}
	h, err := NewSupportChatHandler(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.POST("/chat", h.CreateCustomerConversation)
	form := url.Values{"customer_name": {"ลูกค้าทดสอบ"}}
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.created.CustomerName != "ลูกค้าทดสอบ" {
		t.Fatalf("customer name = %q", repo.created.CustomerName)
	}
	if repo.created.ID == "" || repo.created.CustomerToken == "" || repo.created.ID == repo.created.CustomerToken {
		t.Fatalf("expected independent opaque ids: %#v", repo.created)
	}
}

func TestDeleteCustomerConversationRequiresItsOpaqueToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &supportChatRepoStub{
		created: domain.SupportConversation{ID: "conversation-id", CustomerToken: "secret-token"},
		owned:   true,
	}
	h, err := NewSupportChatHandler(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.DELETE("/chat/:id", h.DeleteCustomerConversation)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/chat/conversation-id", nil)
	req.Header.Set("X-Chat-Token", "secret-token")
	r.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusNoContent || !repo.deleted {
		t.Fatalf("delete status = %d, deleted=%v body=%s", recorder.Code, repo.deleted, recorder.Body.String())
	}

	unauthorizedRepo := &supportChatRepoStub{created: repo.created}
	unauthorizedHandler, err := NewSupportChatHandler(unauthorizedRepo, nil)
	if err != nil {
		t.Fatal(err)
	}
	unauthorizedRouter := gin.New()
	unauthorizedRouter.DELETE("/chat/:id", unauthorizedHandler.DeleteCustomerConversation)
	unauthorizedRecorder := httptest.NewRecorder()
	unauthorizedRequest := httptest.NewRequest(http.MethodDelete, "/chat/conversation-id", nil)
	unauthorizedRequest.Header.Set("X-Chat-Token", "wrong-token")
	unauthorizedRouter.ServeHTTP(unauthorizedRecorder, unauthorizedRequest)
	if unauthorizedRecorder.Code != http.StatusUnauthorized || unauthorizedRepo.deleted {
		t.Fatalf("unauthorized delete status = %d, deleted=%v", unauthorizedRecorder.Code, unauthorizedRepo.deleted)
	}

	missingTokenRecorder := httptest.NewRecorder()
	missingTokenRequest := httptest.NewRequest(http.MethodDelete, "/chat/conversation-id", nil)
	unauthorizedRouter.ServeHTTP(missingTokenRecorder, missingTokenRequest)
	if missingTokenRecorder.Code != http.StatusUnauthorized || unauthorizedRepo.deleted {
		t.Fatalf("missing-token delete status = %d, deleted=%v body=%s", missingTokenRecorder.Code, unauthorizedRepo.deleted, missingTokenRecorder.Body.String())
	}
}

func TestAdminDeleteMessageDeletesRequestedMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &supportChatRepoStub{}
	h, err := NewSupportChatHandler(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.DELETE("/admin/chat/:id/messages/:messageID", h.AdminDeleteMessageJSON)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/admin/chat/conversation-id/messages/message-id", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if repo.deletedMessageID != "message-id" {
		t.Fatalf("deleted message = %q", repo.deletedMessageID)
	}
}

func TestCreateCustomerConversationValidatesAndReturnsAvatar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		inputAvatar    string
		expectedAvatar string
	}{
		{inputAvatar: "cat", expectedAvatar: "cat"},
		{inputAvatar: "dog", expectedAvatar: "dog"},
		{inputAvatar: "rabbit", expectedAvatar: "rabbit"},
		{inputAvatar: "bear", expectedAvatar: "bear"},
		{inputAvatar: "panda", expectedAvatar: "panda"},
		{inputAvatar: "/images/chat_avatar_dog.png", expectedAvatar: "dog"},
		{inputAvatar: "invalid", expectedAvatar: "cat"},
		{inputAvatar: "", expectedAvatar: "cat"},
	}

	for _, tt := range tests {
		repo := &supportChatRepoStub{}
		h, err := NewSupportChatHandler(repo, nil)
		if err != nil {
			t.Fatal(err)
		}
		r := gin.New()
		r.POST("/chat", h.CreateCustomerConversation)
		form := url.Values{
			"customer_name":   {"ลูกค้าทดสอบ"},
			"customer_avatar": {tt.inputAvatar},
		}
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/chat", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		r.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
		}
		if repo.created.CustomerAvatar != tt.expectedAvatar {
			t.Fatalf("input %q expected avatar %q, got %q", tt.inputAvatar, tt.expectedAvatar, repo.created.CustomerAvatar)
		}
		if !strings.Contains(recorder.Body.String(), `"customer_avatar":"`+tt.expectedAvatar+`"`) {
			t.Fatalf("response body should include customer_avatar: %s", recorder.Body.String())
		}
	}
}

func TestCustomerMessagesReturnsAvatarWhenAuthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &supportChatRepoStub{
		created: domain.SupportConversation{
			ID:             "conv-123",
			CustomerToken:  "token-456",
			CustomerName:   "สมหญิง",
			CustomerAvatar: "female",
			Status:         "open",
		},
		owned: true,
	}
	h, err := NewSupportChatHandler(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/chat/:id/messages", h.CustomerMessages)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/chat/conv-123/messages", nil)
	req.Header.Set("X-Chat-Token", "token-456")
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"customer_avatar":"female"`) {
		t.Fatalf("expected customer_avatar in body: %s", recorder.Body.String())
	}
}

func TestCustomerMessagesRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h, err := NewSupportChatHandler(&supportChatRepoStub{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/chat/:id/messages", h.CustomerMessages)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/chat/not-owned/messages", nil))
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
}

func TestCustomerUnreadReturnsCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &supportChatRepoStub{
		created: domain.SupportConversation{
			ID:                  "conv-999",
			CustomerToken:       "token-999",
			CustomerName:        "สมชาย",
			CustomerAvatar:      "cat",
			CustomerUnreadCount: 3,
			Status:              "open",
		},
		owned: true,
	}
	h, err := NewSupportChatHandler(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/chat/:id/unread", h.CustomerUnread)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/chat/conv-999/unread", nil)
	req.Header.Set("X-Chat-Token", "token-999")
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"unread_count":3`) {
		t.Fatalf("expected unread_count 3 in body: %s", recorder.Body.String())
	}
}

func TestRegisterCustomerDeviceToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &supportChatRepoStub{
		created: domain.SupportConversation{
			ID:            "conv-token-1",
			CustomerToken: "token-auth-1",
		},
		owned: true,
	}
	h, err := NewSupportChatHandler(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.POST("/chat/:id/device-token", h.RegisterCustomerDeviceToken)

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/chat/conv-token-1/device-token", strings.NewReader(`{"platform":"android","device_token":"fcm-test-token-123","app_version":"1.3.1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Chat-Token", "token-auth-1")
	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", recorder.Code, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected status ok, got %s", recorder.Body.String())
	}
}
