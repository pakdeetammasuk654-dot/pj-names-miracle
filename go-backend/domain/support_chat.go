package domain

import "time"

// SupportConversation identifies a support thread. CustomerToken is an opaque
// secret held by the client; it is never returned from list or detail calls.
type SupportConversation struct {
	ID                 string    `json:"id"`
	CustomerToken      string    `json:"-"`
	CustomerName       string    `json:"customer_name"`
	CustomerAvatar     string    `json:"customer_avatar"`
	Status              string    `json:"status"`
	AdminUnreadCount    int       `json:"admin_unread_count"`
	CustomerUnreadCount int       `json:"customer_unread_count"`
	LastMessagePreview  string    `json:"last_message_preview"`
	LastSender          string    `json:"last_sender"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type SupportMessage struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Sender         string    `json:"sender"`
	Body           string    `json:"body"`
	ImageURL       string    `json:"image_url"`
	IsRead         bool      `json:"is_read"`
	CreatedAt      time.Time `json:"created_at"`
}

type SupportChatInboxStats struct {
	NeedsReply int `json:"needs_reply"`
	Active     int `json:"active"`
	Closed     int `json:"closed"`
}

type SupportChatDeviceToken struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	Platform       string    `json:"platform"`
	DeviceToken    string    `json:"device_token"`
	AppVersion     string    `json:"app_version"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type RegisterDeviceTokenRequest struct {
	Platform    string `json:"platform" binding:"required"`
	DeviceToken string `json:"device_token" binding:"required"`
	AppVersion  string `json:"app_version"`
}

