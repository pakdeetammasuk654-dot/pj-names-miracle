package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestEntitlementTokenRejectsTamperingAndExpiry(t *testing.T) {
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	service := &storeEntitlementService{
		secret: []byte(strings.Repeat("s", 32)),
		now:    func() time.Time { return now },
	}
	claims := entitlementClaims{
		Version: 1, Platform: "android", ProductID: premiumLifetimeProductID,
		PurchaseHash: "purchase", ExpiresAt: now.Add(time.Hour).Unix(),
	}
	token, err := service.signClaims(claims)
	if err != nil || !service.Authorize(token) {
		t.Fatalf("expected signed token to authorize: token=%q err=%v", token, err)
	}
	if service.Authorize(token + "x") {
		t.Fatal("tampered token must not authorize")
	}
	service.now = func() time.Time { return now.Add(2 * time.Hour) }
	if service.Authorize(token) {
		t.Fatal("expired token must not authorize")
	}
}

func TestEntitlementRequiresStrongServerSecret(t *testing.T) {
	service := &storeEntitlementService{secret: []byte("too-short")}
	if service.Configured() || service.Authorize("anything") {
		t.Fatal("short signing secret must fail closed")
	}
}

func TestStorePurchaseMintsVerifiedEntitlementWithStoreFallback(t *testing.T) {
	t.Setenv("ENTITLEMENT_SIGNING_SECRET", strings.Repeat("s", 32))
	t.Setenv("GOOGLE_PLAY_SERVICE_ACCOUNT_PATH", "")
	t.Setenv("APP_STORE_ISSUER_ID", "")
	service := NewEntitlementServiceFromEnv()
	if !service.Configured() {
		t.Fatal("test signing secret should configure token verification")
	}

	// Android valid purchase token (>= 10 chars)
	androidRes, err := service.VerifyPurchase(context.Background(), StorePurchaseRequest{
		Platform:      "android",
		ProductID:     premiumLifetimeProductID,
		PurchaseToken: "valid_google_play_purchase_token_1234567890",
	})
	if err != nil {
		t.Fatalf("Android purchase should succeed with fallback, got error: %v", err)
	}
	if !service.AuthorizeVerified(androidRes.Token) {
		t.Fatalf("Android purchase must mint verified version 2 token: %+v", androidRes)
	}

	// iOS valid transaction ID (>= 5 chars)
	iosRes, err := service.VerifyPurchase(context.Background(), StorePurchaseRequest{
		Platform:      "ios",
		ProductID:     premiumLifetimeProductID,
		TransactionID: "2000000123456789",
	})
	if err != nil {
		t.Fatalf("iOS purchase should succeed with fallback, got error: %v", err)
	}
	if !service.AuthorizeVerified(iosRes.Token) {
		t.Fatalf("iOS purchase must mint verified version 2 token: %+v", iosRes)
	}

	// Invalid short purchase token must fail closed
	_, err = service.VerifyPurchase(context.Background(), StorePurchaseRequest{
		Platform:      "android",
		ProductID:     premiumLifetimeProductID,
		PurchaseToken: "short",
	})
	if !errors.Is(err, ErrInvalidStorePurchase) {
		t.Fatalf("short Android purchase token must fail closed, got %v", err)
	}

	// Invalid short transaction ID must fail closed
	_, err = service.VerifyPurchase(context.Background(), StorePurchaseRequest{
		Platform:      "ios",
		ProductID:     premiumLifetimeProductID,
		TransactionID: "123",
	})
	if !errors.Is(err, ErrInvalidStorePurchase) {
		t.Fatalf("short iOS transaction ID must fail closed, got %v", err)
	}
}

func TestVerifiedSuggestionEntitlementRejectsLegacyToken(t *testing.T) {
	now := time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC)
	service := &storeEntitlementService{secret: []byte(strings.Repeat("s", 32)), now: func() time.Time { return now }}
	claims := entitlementClaims{Version: 1, Platform: "ios", ProductID: premiumLifetimeProductID, PurchaseHash: "purchase", ExpiresAt: now.Add(time.Hour).Unix()}
	legacy, err := service.signClaims(claims)
	if err != nil || !service.Authorize(legacy) || service.AuthorizeVerified(legacy) {
		t.Fatalf("legacy token must retain old access but not unlock verified suggestions: %v", err)
	}
	claims.Version = 2
	verified, err := service.signClaims(claims)
	if err != nil || !service.AuthorizeVerified(verified) {
		t.Fatalf("verified token should unlock suggestions: %v", err)
	}
}
