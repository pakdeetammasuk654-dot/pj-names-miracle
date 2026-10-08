package usecase

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"go-backend/domain"
	"go-backend/repository"
)

type PushNotificationService interface {
	SendAdminReplyNotification(ctx context.Context, conversationID string, message domain.SupportMessage)
	SendCustomerMessageNotification(ctx context.Context, conversationID string, message domain.SupportMessage)
}

type pushNotificationService struct {
	chatRepo repository.SupportChatRepository
	client   *http.Client

	apnsKeyMu   sync.Mutex
	apnsPrivKey *ecdsa.PrivateKey
	apnsToken   string
	apnsTokenAt time.Time

	googleTokenMu sync.Mutex
	googleToken   string
	googleTokenAt time.Time
}

func NewPushNotificationService(chatRepo repository.SupportChatRepository) PushNotificationService {
	return &pushNotificationService{
		chatRepo: chatRepo,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type PushPayload struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Sender         string `json:"sender"`
	Badge          int    `json:"badge,omitempty"`
}

func (s *pushNotificationService) SendAdminReplyNotification(ctx context.Context, conversationID string, message domain.SupportMessage) {
	tokens, err := s.chatRepo.GetActiveDeviceTokens(ctx, conversationID)
	if err != nil {
		log.Printf("[PushNotification] Error fetching tokens for conversation %s: %v", conversationID, err)
		return
	}
	if len(tokens) == 0 {
		return
	}

	bodyPreview := strings.TrimSpace(message.Body)
	if bodyPreview == "" && message.ImageURL != "" {
		bodyPreview = "📷 ส่งรูปภาพประกอบ"
	}
	if len([]rune(bodyPreview)) > 120 {
		bodyPreview = string([]rune(bodyPreview)[:120]) + "..."
	}

	badgeCount := 1
	conv, err := s.chatRepo.GetConversation(ctx, conversationID)
	if err == nil && conv != nil && conv.CustomerUnreadCount > 0 {
		badgeCount = conv.CustomerUnreadCount
	}

	payload := PushPayload{
		Title:          "อ.ทญา ว่องไว",
		Body:           bodyPreview,
		ConversationID: conversationID,
		MessageID:      message.ID,
		Sender:         "admin",
		Badge:          badgeCount,
	}

	for _, token := range tokens {
		go func(t domain.SupportChatDeviceToken) {
			bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			switch strings.ToLower(t.Platform) {
			case "android":
				s.sendFCM(bgCtx, t, payload)
			case "ios":
				s.sendAPNs(bgCtx, t, payload)
			case "web":
				s.sendWebPush(bgCtx, t, payload)
			}
		}(token)
	}
}

type googleServiceAccount struct {
	Type        string `json:"type"`
	ProjectID   string `json:"project_id"`
	PrivateKey  string `json:"private_key"`
	ClientEmail string `json:"client_email"`
	TokenURI    string `json:"token_uri"`
}

func (s *pushNotificationService) sendFCM(ctx context.Context, t domain.SupportChatDeviceToken, p PushPayload) {
	saPath := os.Getenv("FCM_SERVICE_ACCOUNT_PATH")
	projectID := os.Getenv("FCM_PROJECT_ID")
	if projectID == "" {
		projectID = "miracle-names"
	}

	if saPath != "" {
		accessToken, err := s.getGoogleAccessToken(saPath)
		if err == nil && accessToken != "" {
			s.sendFCMv1(ctx, t, p, projectID, accessToken)
			return
		} else if err != nil {
			log.Printf("[PushNotification] FCM v1 auth error: %v", err)
		}
	}

	serverKey := os.Getenv("FCM_SERVER_KEY")
	if serverKey == "" {
		log.Printf("[PushNotification] FCM skipped for Android device (FCM credentials not configured)")
		return
	}

	reqBody := map[string]any{
		"to": t.DeviceToken,
		"notification": map[string]string{
			"title": p.Title,
			"body":  p.Body,
			"sound": "default",
		},
		"data": map[string]string{
			"conversation_id": p.ConversationID,
			"message_id":      p.MessageID,
			"sender":          p.Sender,
			"click_action":    "OPEN_SUPPORT_CHAT",
		},
		"priority": "high",
	}

	jsonBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "key="+serverKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("[PushNotification] FCM Legacy send error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("[PushNotification] FCM Legacy sent successfully to conversation=%s", t.ConversationID)
	} else {
		log.Printf("[PushNotification] FCM Legacy responded with status=%d", resp.StatusCode)
	}
}

func (s *pushNotificationService) sendFCMv1(ctx context.Context, t domain.SupportChatDeviceToken, p PushPayload, projectID, accessToken string) {
	urlEndpoint := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", projectID)

	reqBody := map[string]any{
		"message": map[string]any{
			"token": t.DeviceToken,
			"notification": map[string]string{
				"title": p.Title,
				"body":  p.Body,
			},
			"data": map[string]string{
				"conversation_id": p.ConversationID,
				"message_id":      p.MessageID,
				"sender":          p.Sender,
				"title":           p.Title,
				"body":            p.Body,
			},
			"android": map[string]any{
				"priority": "HIGH",
				"notification": map[string]string{
					"channel_id":   "chuedee_support_chat",
					"icon":         "ic_notification_chuedee",
					"color":        "#D4AF37",
					"sound":        "default",
					"click_action": "OPEN_SUPPORT_CHAT",
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlEndpoint, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("[PushNotification] FCM v1 send error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("[PushNotification] FCM v1 sent successfully to conversation=%s", t.ConversationID)
	} else {
		bodyResp, _ := io.ReadAll(resp.Body)
		log.Printf("[PushNotification] FCM v1 responded with status=%d, body=%s", resp.StatusCode, string(bodyResp))
		if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone || strings.Contains(string(bodyResp), "UNREGISTERED") {
			_ = s.chatRepo.DeactivateDeviceToken(context.Background(), t.DeviceToken)
		}
	}
}

func (s *pushNotificationService) getGoogleAccessToken(saPath string) (string, error) {
	s.googleTokenMu.Lock()
	defer s.googleTokenMu.Unlock()

	if s.googleToken != "" && time.Since(s.googleTokenAt) < 50*time.Minute {
		return s.googleToken, nil
	}

	data, err := os.ReadFile(saPath)
	if err != nil {
		return "", fmt.Errorf("read service account file: %w", err)
	}

	var sa googleServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return "", fmt.Errorf("unmarshal service account JSON: %w", err)
	}

	block, _ := pem.Decode([]byte(sa.PrivateKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM from private key")
	}

	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	rsaKey, ok := parsedKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not an RSA private key")
	}

	headerJSON, _ := json.Marshal(map[string]string{"alg": "RS256", "typ": "JWT"})
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	now := time.Now().Unix()
	claimsJSON, _ := json.Marshal(map[string]any{
		"iss":   sa.ClientEmail,
		"scope": "https://www.googleapis.com/auth/firebase.messaging",
		"aud":   "https://oauth2.googleapis.com/token",
		"exp":   now + 3600,
		"iat":   now,
	})
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("sign RSA jwt: %w", err)
	}

	sigB64 := base64.RawURLEncoding.EncodeToString(signature)
	signedJWT := signingInput + "." + sigB64

	tokenURI := sa.TokenURI
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}

	form := url.Values{}
	form.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	form.Set("assertion", signedJWT)

	resp, err := s.client.PostForm(tokenURI, form)
	if err != nil {
		return "", fmt.Errorf("exchange JWT for access token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	s.googleToken = tokenResp.AccessToken
	s.googleTokenAt = time.Now()
	return s.googleToken, nil
}

func (s *pushNotificationService) sendAPNs(ctx context.Context, t domain.SupportChatDeviceToken, p PushPayload) {
	keyPath := os.Getenv("APNS_KEY_PATH")
	keyID := os.Getenv("APNS_KEY_ID")
	teamID := os.Getenv("APNS_TEAM_ID")
	bundleID := os.Getenv("APNS_BUNDLE_ID")
	if bundleID == "" {
		bundleID = "com.numberniceic.app"
	}
	if strings.ToLower(t.Platform) == "ios_admin" || t.ConversationID == "admin" {
		adminBundleID := os.Getenv("APNS_ADMIN_BUNDLE_ID")
		if adminBundleID == "" {
			adminBundleID = "com.numberniceic.admin"
		}
		bundleID = adminBundleID
	}
	if keyPath == "" || keyID == "" || teamID == "" {
		log.Printf("[PushNotification] APNs skipped for iOS device (APNS credentials not configured)")
		return
	}

	authToken, err := s.getAPNsAuthToken(keyPath, keyID, teamID)
	if err != nil {
		log.Printf("[PushNotification] APNs auth token error: %v", err)
		return
	}

	hosts := []string{"https://api.push.apple.com", "https://api.sandbox.push.apple.com"}
	if os.Getenv("APNS_ENV") == "sandbox" || os.Getenv("APNS_ENV") == "development" {
		hosts = []string{"https://api.sandbox.push.apple.com", "https://api.push.apple.com"}
	}

	badgeCount := p.Badge
	if badgeCount <= 0 {
		badgeCount = 1
	}

	apnsPayload := map[string]any{
		"aps": map[string]any{
			"alert": map[string]string{
				"title": p.Title,
				"body":  p.Body,
			},
			"sound": "default",
			"badge": badgeCount,
		},
		"conversation_id": p.ConversationID,
		"message_id":      p.MessageID,
		"sender":          p.Sender,
	}

	jsonBytes, err := json.Marshal(apnsPayload)
	if err != nil {
		return
	}

	for _, host := range hosts {
		apnsURL := fmt.Sprintf("%s/3/device/%s", host, t.DeviceToken)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, apnsURL, bytes.NewBuffer(jsonBytes))
		if err != nil {
			continue
		}
		req.Header.Set("Authorization", "bearer "+authToken)
		req.Header.Set("apns-topic", bundleID)
		req.Header.Set("apns-push-type", "alert")
		req.Header.Set("apns-priority", "10")
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("[PushNotification] APNs send error on %s: %v", host, err)
			continue
		}

		bodyResp, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("[PushNotification] APNs sent successfully on %s to conversation=%s", host, t.ConversationID)
			return
		} else if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
			log.Printf("[PushNotification] APNs token expired on %s, deactivating: %s", host, t.DeviceToken)
			_ = s.chatRepo.DeactivateDeviceToken(context.Background(), t.DeviceToken)
			return
		} else if resp.StatusCode == http.StatusBadRequest && strings.Contains(string(bodyResp), "BadDeviceToken") {
			log.Printf("[PushNotification] APNs host %s returned BadDeviceToken, trying fallback host...", host)
			continue
		} else {
			log.Printf("[PushNotification] APNs responded with status=%d, body=%s for conv=%s on %s", resp.StatusCode, string(bodyResp), t.ConversationID, host)
		}
	}
}

func (s *pushNotificationService) getAPNsAuthToken(keyPath, keyID, teamID string) (string, error) {
	s.apnsKeyMu.Lock()
	defer s.apnsKeyMu.Unlock()

	if s.apnsToken != "" && time.Since(s.apnsTokenAt) < 45*time.Minute {
		return s.apnsToken, nil
	}

	if s.apnsPrivKey == nil {
		keyBytes, err := os.ReadFile(keyPath)
		if err != nil {
			return "", fmt.Errorf("read apns key file: %w", err)
		}
		block, _ := pem.Decode(keyBytes)
		if block == nil {
			return "", fmt.Errorf("invalid PEM in apns key file")
		}
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", fmt.Errorf("parse PKCS8 apns private key: %w", err)
		}
		ecdsaKey, ok := parsedKey.(*ecdsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("apns key is not an ECDSA private key")
		}
		s.apnsPrivKey = ecdsaKey
	}

	header := map[string]string{
		"alg": "ES256",
		"kid": keyID,
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	now := time.Now().Unix()
	claims := map[string]any{
		"iss": teamID,
		"iat": now,
	}
	claimsJSON, _ := json.Marshal(claims)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hash := hasher.Sum(nil)

	r, sBig, err := ecdsa.Sign(rand.Reader, s.apnsPrivKey, hash)
	if err != nil {
		return "", fmt.Errorf("sign apns jwt: %w", err)
	}

	curveBits := s.apnsPrivKey.Curve.Params().BitSize
	keyBytesLen := (curveBits + 7) / 8
	sig := make([]byte, 2*keyBytesLen)
	rBytes := r.Bytes()
	sBytes := sBig.Bytes()
	copy(sig[keyBytesLen-len(rBytes):keyBytesLen], rBytes)
	copy(sig[2*keyBytesLen-len(sBytes):], sBytes)

	sigB64 := base64.RawURLEncoding.EncodeToString(sig)
	token := signingInput + "." + sigB64

	s.apnsToken = token
	s.apnsTokenAt = time.Now()
	return token, nil
}

func (s *pushNotificationService) sendWebPush(ctx context.Context, t domain.SupportChatDeviceToken, p PushPayload) {
	log.Printf("[PushNotification] Web push dispatch prepared for conv=%s", t.ConversationID)
}

func (s *pushNotificationService) SendCustomerMessageNotification(ctx context.Context, conversationID string, message domain.SupportMessage) {
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var customerName string
		conv, err := s.chatRepo.GetConversation(bgCtx, conversationID)
		if err == nil && conv != nil && strings.TrimSpace(conv.CustomerName) != "" {
			customerName = strings.TrimSpace(conv.CustomerName)
		} else {
			customerName = "ลูกค้า"
		}

		bodyPreview := strings.TrimSpace(message.Body)
		if bodyPreview == "" && message.ImageURL != "" {
			bodyPreview = "📷 แนบรูปภาพประกอบ"
		}
		if len([]rune(bodyPreview)) > 120 {
			bodyPreview = string([]rune(bodyPreview)[:120]) + "..."
		}

		badgeCount := 1
		stats, err := s.chatRepo.GetInboxStats(bgCtx)
		if err == nil && stats.NeedsReply > 0 {
			badgeCount = stats.NeedsReply
		}

		payload := PushPayload{
			Title:          fmt.Sprintf("💬 คุณ %s", customerName),
			Body:           bodyPreview,
			ConversationID: conversationID,
			MessageID:      message.ID,
			Sender:         "customer",
			Badge:          badgeCount,
		}

		// 1. Dispatch APNs/FCM to registered Admin devices
		adminTokens, err := s.chatRepo.GetActiveAdminDeviceTokens(bgCtx)
		if err == nil && len(adminTokens) > 0 {
			for _, t := range adminTokens {
				switch strings.ToLower(t.Platform) {
				case "ios", "ios_admin":
					s.sendAPNs(bgCtx, t, payload)
				case "android", "android_admin":
					s.sendFCM(bgCtx, t, payload)
				}
			}
		}

		// 2. Dispatch to Telegram Bot (if configured)
		botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
		chatID := os.Getenv("TELEGRAM_CHAT_ID")
		if botToken != "" && chatID != "" {
			text := fmt.Sprintf("💬 *มีข้อความใหม่จากลูกค้า!*\n👤 *ลูกค้า:* คุณ %s\n📝 *ข้อความ:* %s", customerName, bodyPreview)
			chatURL := fmt.Sprintf("https://xn--b3cu8e7ah6h.com/admin/chat/%s?view=active", conversationID)

			reqBody := map[string]any{
				"chat_id":    chatID,
				"text":       text,
				"parse_mode": "Markdown",
				"reply_markup": map[string]any{
					"inline_keyboard": [][]map[string]string{
						{
							{"text": "👉 เปิดตอบแชตทันที", "url": chatURL},
						},
					},
				},
			}

			jsonBytes, err := json.Marshal(reqBody)
			if err == nil {
				apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
				req, err := http.NewRequestWithContext(bgCtx, http.MethodPost, apiURL, bytes.NewBuffer(jsonBytes))
				if err == nil {
					req.Header.Set("Content-Type", "application/json")
					resp, err := s.client.Do(req)
					if err == nil {
						resp.Body.Close()
					}
				}
			}
		}
	}()
}

var _ = big.NewInt
