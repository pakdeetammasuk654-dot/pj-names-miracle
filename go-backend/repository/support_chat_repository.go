package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go-backend/domain"
)

type SupportChatRepository interface {
	CreateConversation(context.Context, domain.SupportConversation) error
	GetConversation(context.Context, string) (*domain.SupportConversation, error)
	GetConversationForCustomer(context.Context, string, string) (*domain.SupportConversation, error)
	ListCustomerImageURLs(context.Context, string, string) ([]string, error)
	DeleteConversationForCustomer(context.Context, string, string) error
	ListConversations(context.Context, string) ([]domain.SupportConversation, error)
	ListMessages(context.Context, string) ([]domain.SupportMessage, error)
	CreateMessage(context.Context, domain.SupportMessage) error
	DeleteMessage(context.Context, string, string) (string, error)
	MarkCustomerRead(context.Context, string) error
	MarkAdminRead(context.Context, string) error
	SetConversationStatus(context.Context, string, string) error
	GetInboxStats(context.Context) (domain.SupportChatInboxStats, error)
	RegisterDeviceToken(ctx context.Context, conversationID, platform, token, appVersion string) error
	GetActiveDeviceTokens(ctx context.Context, conversationID string) ([]domain.SupportChatDeviceToken, error)
	DeactivateDeviceToken(ctx context.Context, token string) error
	RegisterAdminDeviceToken(ctx context.Context, platform, token, appVersion string) error
	GetActiveAdminDeviceTokens(ctx context.Context) ([]domain.SupportChatDeviceToken, error)
	DeactivateAdminDeviceToken(ctx context.Context, token string) error
}

type supportChatRepository struct{ db *sql.DB }

func NewSupportChatRepository(db *sql.DB) SupportChatRepository {
	return &supportChatRepository{db: db}
}

func (r *supportChatRepository) CreateConversation(ctx context.Context, c domain.SupportConversation) error {
	avatar := c.CustomerAvatar
	if avatar == "" {
		avatar = "cat"
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO support_conversations (id, customer_token, customer_name, customer_avatar) VALUES ($1,$2,$3,$4)`, c.ID, c.CustomerToken, c.CustomerName, avatar)
	return err
}

func (r *supportChatRepository) GetConversation(ctx context.Context, id string) (*domain.SupportConversation, error) {
	return scanSupportConversation(r.db.QueryRowContext(ctx, `SELECT id::text, customer_token::text, customer_name, customer_avatar, status, admin_unread_count, customer_unread_count, last_message_preview, last_sender, created_at, updated_at FROM support_conversations WHERE id=$1`, id))
}

func (r *supportChatRepository) GetConversationForCustomer(ctx context.Context, id, token string) (*domain.SupportConversation, error) {
	return scanSupportConversation(r.db.QueryRowContext(ctx, `SELECT id::text, customer_token::text, customer_name, customer_avatar, status, admin_unread_count, customer_unread_count, last_message_preview, last_sender, created_at, updated_at FROM support_conversations WHERE id=$1 AND customer_token=$2`, id, token))
}

func (r *supportChatRepository) ListCustomerImageURLs(ctx context.Context, id, token string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.image_url
		FROM support_messages AS m
		JOIN support_conversations AS c ON c.id = m.conversation_id
		WHERE c.id = $1 AND c.customer_token = $2 AND m.image_url <> ''
	`, id, token)
	if err != nil {
		return nil, fmt.Errorf("list customer chat images: %w", err)
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var imageURL string
		if err := rows.Scan(&imageURL); err != nil {
			return nil, err
		}
		urls = append(urls, imageURL)
	}
	return urls, rows.Err()
}

func (r *supportChatRepository) DeleteConversationForCustomer(ctx context.Context, id, token string) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM support_conversations WHERE id=$1 AND customer_token=$2`, id, token)
	if err != nil {
		return fmt.Errorf("delete customer chat: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if deleted == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *supportChatRepository) ListConversations(ctx context.Context, inbox string) ([]domain.SupportConversation, error) {
	where := ""
	switch inbox {
	case "needs_reply":
		where = "WHERE status = 'open' AND admin_unread_count > 0"
	case "active":
		where = "WHERE status = 'open' AND admin_unread_count = 0"
	case "closed":
		where = "WHERE status = 'closed'"
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id::text, customer_token::text, customer_name, customer_avatar, status, admin_unread_count, customer_unread_count, last_message_preview, last_sender, created_at, updated_at FROM support_conversations `+where+` ORDER BY admin_unread_count DESC, updated_at DESC LIMIT 50`)
	if err != nil {
		return nil, fmt.Errorf("list support conversations: %w", err)
	}
	defer rows.Close()
	items := make([]domain.SupportConversation, 0)
	for rows.Next() {
		item, err := scanSupportConversation(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *supportChatRepository) ListMessages(ctx context.Context, conversationID string) ([]domain.SupportMessage, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id::text, conversation_id::text, sender, body, image_url, is_read, created_at FROM support_messages WHERE conversation_id=$1 ORDER BY created_at ASC`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("list support messages: %w", err)
	}
	defer rows.Close()
	items := make([]domain.SupportMessage, 0)
	for rows.Next() {
		var item domain.SupportMessage
		if err := rows.Scan(&item.ID, &item.ConversationID, &item.Sender, &item.Body, &item.ImageURL, &item.IsRead, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *supportChatRepository) CreateMessage(ctx context.Context, m domain.SupportMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO support_messages (id, conversation_id, sender, body, image_url) VALUES ($1,$2,$3,$4,$5)`, m.ID, m.ConversationID, m.Sender, m.Body, m.ImageURL); err != nil {
		return err
	}
	preview := m.Body
	if preview == "" {
		preview = "[รูปภาพ]"
	}
	if len([]rune(preview)) > 140 {
		preview = string([]rune(preview)[:140]) + "…"
	}
	if _, err = tx.ExecContext(ctx, `UPDATE support_conversations
		SET updated_at=NOW(), last_message_preview=$2, last_sender=$3,
		admin_unread_count=CASE WHEN $3='customer' THEN admin_unread_count+1 ELSE 0 END,
		customer_unread_count=CASE WHEN $3='admin' THEN customer_unread_count+1 ELSE 0 END,
		status=CASE WHEN $3='customer' THEN 'open' ELSE status END
		WHERE id=$1`, m.ConversationID, preview, m.Sender); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteMessage removes one message belonging to a conversation and rebuilds
// the conversation summary and unread counts from the remaining messages.
// It returns the deleted image URL so the caller can remove its stored file.
func (r *supportChatRepository) DeleteMessage(ctx context.Context, conversationID, messageID string) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var imageURL string
	err = tx.QueryRowContext(ctx, `
		DELETE FROM support_messages
		WHERE id=$1 AND conversation_id=$2
		RETURNING image_url
	`, messageID, conversationID).Scan(&imageURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("delete support message: %w", err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE support_conversations AS c
		SET
			updated_at = NOW(),
			last_message_preview = COALESCE((
				SELECT CASE WHEN m.body <> '' THEN
					CASE WHEN char_length(m.body) > 140 THEN left(m.body, 140) || '…' ELSE m.body END
					ELSE '[รูปภาพ]' END
				FROM support_messages AS m
				WHERE m.conversation_id = c.id
				ORDER BY m.created_at DESC, m.id DESC
				LIMIT 1
			), ''),
			last_sender = COALESCE((
				SELECT m.sender FROM support_messages AS m
				WHERE m.conversation_id = c.id
				ORDER BY m.created_at DESC, m.id DESC
				LIMIT 1
			), ''),
			admin_unread_count = (
				SELECT COUNT(*) FROM support_messages AS m
				WHERE m.conversation_id = c.id AND m.sender = 'customer' AND m.is_read = FALSE
			),
			customer_unread_count = (
				SELECT COUNT(*) FROM support_messages AS m
				WHERE m.conversation_id = c.id AND m.sender = 'admin' AND m.is_read = FALSE
			)
		WHERE c.id = $1
	`, conversationID)
	if err != nil {
		return "", fmt.Errorf("update support conversation after message deletion: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return imageURL, nil
}

func (r *supportChatRepository) MarkCustomerRead(ctx context.Context, conversationID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE support_messages SET is_read=true WHERE conversation_id=$1 AND sender='admin' AND is_read=false`, conversationID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE support_conversations SET customer_unread_count=0 WHERE id=$1`, conversationID)
	return err
}

func (r *supportChatRepository) MarkAdminRead(ctx context.Context, conversationID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE support_messages SET is_read=true WHERE conversation_id=$1 AND sender='customer' AND is_read=false`, conversationID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `UPDATE support_conversations SET admin_unread_count=0 WHERE id=$1`, conversationID)
	return err
}

func (r *supportChatRepository) SetConversationStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE support_conversations SET status=$2, updated_at=NOW() WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	if changed, err := result.RowsAffected(); err != nil {
		return err
	} else if changed == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *supportChatRepository) GetInboxStats(ctx context.Context) (domain.SupportChatInboxStats, error) {
	var stats domain.SupportChatInboxStats
	err := r.db.QueryRowContext(ctx, `SELECT
		COUNT(*) FILTER (WHERE status='open' AND admin_unread_count > 0),
		COUNT(*) FILTER (WHERE status='open' AND admin_unread_count = 0),
		COUNT(*) FILTER (WHERE status='closed')
		FROM support_conversations`).Scan(&stats.NeedsReply, &stats.Active, &stats.Closed)
	if err != nil {
		return domain.SupportChatInboxStats{}, fmt.Errorf("get support inbox stats: %w", err)
	}
	return stats, nil
}

func (r *supportChatRepository) RegisterDeviceToken(ctx context.Context, conversationID, platform, token, appVersion string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO support_chat_device_tokens (conversation_id, platform, device_token, app_version, is_active, updated_at)
		VALUES ($1, $2, $3, $4, TRUE, NOW())
		ON CONFLICT (conversation_id, device_token)
		DO UPDATE SET platform=EXCLUDED.platform, app_version=EXCLUDED.app_version, is_active=TRUE, updated_at=NOW()
	`, conversationID, platform, token, appVersion)
	return err
}

func (r *supportChatRepository) GetActiveDeviceTokens(ctx context.Context, conversationID string) ([]domain.SupportChatDeviceToken, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, conversation_id::text, platform, device_token, app_version, is_active, created_at, updated_at
		FROM support_chat_device_tokens
		WHERE conversation_id = $1 AND is_active = TRUE
	`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("get active device tokens: %w", err)
	}
	defer rows.Close()

	var tokens []domain.SupportChatDeviceToken
	for rows.Next() {
		var t domain.SupportChatDeviceToken
		if err := rows.Scan(&t.ID, &t.ConversationID, &t.Platform, &t.DeviceToken, &t.AppVersion, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

func (r *supportChatRepository) DeactivateDeviceToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE support_chat_device_tokens SET is_active=FALSE, updated_at=NOW() WHERE device_token=$1`, token)
	return err
}

func (r *supportChatRepository) RegisterAdminDeviceToken(ctx context.Context, platform, token, appVersion string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_device_tokens (platform, device_token, app_version, is_active, updated_at)
		VALUES ($1, $2, $3, TRUE, NOW())
		ON CONFLICT (device_token)
		DO UPDATE SET platform = EXCLUDED.platform, app_version = EXCLUDED.app_version, is_active = TRUE, updated_at = NOW()
	`, platform, token, appVersion)
	return err
}

func (r *supportChatRepository) GetActiveAdminDeviceTokens(ctx context.Context) ([]domain.SupportChatDeviceToken, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, 'admin' AS conversation_id, platform, device_token, app_version, is_active, created_at, updated_at
		FROM admin_device_tokens
		WHERE is_active = TRUE
	`)
	if err != nil {
		return nil, fmt.Errorf("get active admin device tokens: %w", err)
	}
	defer rows.Close()

	var tokens []domain.SupportChatDeviceToken
	for rows.Next() {
		var t domain.SupportChatDeviceToken
		if err := rows.Scan(&t.ID, &t.ConversationID, &t.Platform, &t.DeviceToken, &t.AppVersion, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, rows.Err()
}

func (r *supportChatRepository) DeactivateAdminDeviceToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE admin_device_tokens SET is_active=FALSE, updated_at=NOW() WHERE device_token=$1`, token)
	return err
}

type supportChatScanner interface{ Scan(...any) error }

func scanSupportConversation(s supportChatScanner) (*domain.SupportConversation, error) {
	var item domain.SupportConversation
	if err := s.Scan(&item.ID, &item.CustomerToken, &item.CustomerName, &item.CustomerAvatar, &item.Status, &item.AdminUnreadCount, &item.CustomerUnreadCount, &item.LastMessagePreview, &item.LastSender, &item.CreatedAt, &item.UpdatedAt); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &item, nil
}
