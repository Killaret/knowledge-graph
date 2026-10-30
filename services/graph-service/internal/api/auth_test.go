package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"knowledge-graph-graph-service/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

func generateTestToken(userID, secret, tokenType string) string {
	claims := TokenClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}

func contextDumpHandler(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	public := isPublicRequest(r.Context())
	skip := isSkipAuthRequest(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"user_id":"` + uid + `","public":` + boolString(public) + `,"skip_auth":` + boolString(skip) + `}`))
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestAuthMiddlewarePrivateRequiresAuth(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
	handler := AuthMiddleware(cfg, false, contextDumpHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestAuthMiddlewareAcceptsValidJWT(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
	handler := AuthMiddleware(cfg, false, contextDumpHandler)

	token := generateTestToken("user-123", cfg.JWTSecret, "access")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !contains(rr.Body.String(), `"user_id":"user-123"`) {
		t.Fatalf("expected user_id in response, got %s", rr.Body.String())
	}
}

func TestAuthMiddlewareRejectsWrongTokenType(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
	handler := AuthMiddleware(cfg, false, contextDumpHandler)

	token := generateTestToken("user-123", cfg.JWTSecret, "refresh")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh token, got %d", rr.Code)
	}
}

func TestAuthMiddlewareInternalTokenWithUserID(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", TrustUserHeader: true, SkipAuth: false}
	handler := AuthMiddleware(cfg, false, contextDumpHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	req.Header.Set("X-Internal-Auth", cfg.InternalAuthToken)
	req.Header.Set("X-User-Id", "internal-user")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !contains(rr.Body.String(), `"user_id":"internal-user"`) {
		t.Fatalf("expected internal user_id, got %s", rr.Body.String())
	}
}

func TestAuthMiddlewareIgnoresInternalUserHeaderByDefault(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
	handler := AuthMiddleware(cfg, false, contextDumpHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	req.Header.Set("X-Internal-Auth", cfg.InternalAuthToken)
	req.Header.Set("X-User-Id", "forged-user")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when trust is disabled, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestGRPCAuthIgnoresInternalUserHeaderByDefault(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-internal-auth", cfg.InternalAuthToken,
		"x-user-id", "forged-user",
	))

	userID, ok := grpcAuth(ctx, cfg)
	if ok || userID != "" {
		t.Fatalf("expected untrusted user header to be rejected, got user_id=%q ok=%v", userID, ok)
	}
}

func TestGRPCAuthAcceptsInternalUserHeaderWhenEnabled(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", TrustUserHeader: true, SkipAuth: false}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"x-internal-auth", cfg.InternalAuthToken,
		"x-user-id", "internal-user",
	))

	userID, ok := grpcAuth(ctx, cfg)
	if !ok || userID != "internal-user" {
		t.Fatalf("expected trusted internal user, got user_id=%q ok=%v", userID, ok)
	}
}

func TestAuthMiddlewareInternalTokenWithoutUserIDRejected(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
	handler := AuthMiddleware(cfg, false, contextDumpHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	req.Header.Set("X-Internal-Auth", cfg.InternalAuthToken)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without user id, got %d", rr.Code)
	}
}

func TestAuthMiddlewarePublicEndpointSkipsAuth(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: false}
	handler := AuthMiddleware(cfg, true, contextDumpHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/public", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !contains(rr.Body.String(), `"public":true`) {
		t.Fatalf("expected public=true, got %s", rr.Body.String())
	}
}

func TestAuthMiddlewareSkipAuthAllowsAll(t *testing.T) {
	cfg := &config.Config{JWTSecret: "test-secret", InternalAuthToken: "internal", SkipAuth: true}
	handler := AuthMiddleware(cfg, false, contextDumpHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/graph/full", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if !contains(body, `"skip_auth":true`) {
		t.Fatalf("expected SKIP_AUTH to be marked in context, got %s", body)
	}
	if contains(body, `"public":true`) {
		t.Fatalf("expected SKIP_AUTH not to be treated as public-only, got %s", body)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || containsInternal(s, substr))
}

func containsInternal(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
