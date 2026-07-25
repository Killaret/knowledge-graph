package api

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"knowledge-graph-graph-service/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TokenClaims represents the expected JWT claims. Both user-facing access tokens
// and internal service tokens use the same shape; the secret used for validation
// differs.
type TokenClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// AuthMiddleware returns an http.HandlerFunc wrapper that authenticates the
// request before passing it to next. Public endpoints (public=true) skip auth
// and mark the request context as public. When SKIP_AUTH is enabled, every
// request is allowed and user_id is set to "public".
func AuthMiddleware(cfg *config.Config, public bool, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.SkipAuth {
			log.Printf("[Auth] SKIP_AUTH enabled, request allowed as user_id=public")
			next(w, r.WithContext(withUserID(r.Context(), "")))
			return
		}

		if public {
			next(w, r.WithContext(withPublic(r.Context())))
			return
		}

		userID, ok := authenticateRequest(r, cfg)
		if !ok || userID == "" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		next(w, r.WithContext(withUserID(r.Context(), userID)))
	}
}

func authenticateRequest(r *http.Request, cfg *config.Config) (string, bool) {
	if authz := r.Header.Get("Authorization"); authz != "" {
		const prefix = "Bearer "
		if strings.HasPrefix(authz, prefix) {
			token := strings.TrimPrefix(authz, prefix)
			return validateJWT(token, cfg.JWTSecret)
		}
	}

	if internal := r.Header.Get("X-Internal-Auth"); internal != "" {
		return validateInternalAuthWithRequest(r, internal, cfg)
	}

	return "", false
}

func validateJWT(tokenString, secret string) (string, bool) {
	if secret == "" {
		return "", false
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithLeeway(5*time.Second))
	if err != nil {
		log.Printf("[Auth] JWT parse error: %v", err)
		return "", false
	}
	if !token.Valid {
		return "", false
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return "", false
	}
	if claims.TokenType != "access" {
		return "", false
	}

	if claims.UserID != "" {
		return claims.UserID, true
	}
	if claims.Subject != "" {
		return claims.Subject, true
	}
	return "", false
}

func validateInternalAuthWithRequest(r *http.Request, token string, cfg *config.Config) (string, bool) {
	if cfg.InternalAuthToken == "" {
		return "", false
	}

	// Exact match with the configured internal token means the request is coming
	// from a trusted internal proxy. If an X-User-Id header is also present,
	// trust it for server-to-server calls; otherwise a private endpoint must
	// also supply a signed token with a user_id.
	if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.InternalAuthToken)) == 1 {
		if userID := r.Header.Get("X-User-Id"); userID != "" {
			return userID, true
		}
		return "", true
	}

	// Alternatively the header may contain a JWT signed with the internal token
	// and carrying a concrete user_id for server-to-server calls.
	return validateJWT(token, cfg.InternalAuthToken)
}

func validateInternalAuth(token string, cfg *config.Config) (string, bool) {
	if cfg.InternalAuthToken == "" {
		return "", false
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(cfg.InternalAuthToken)) == 1 {
		return "", true
	}
	return validateJWT(token, cfg.InternalAuthToken)
}

// grpcAuth extracts a user ID from gRPC metadata using the same rules as the
// HTTP middleware: Authorization Bearer JWT, or signed X-Internal-Auth.
func grpcAuth(ctx context.Context, cfg *config.Config) (string, bool) {
	if cfg.SkipAuth {
		return "", true
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", false
	}

	if vals := md.Get("authorization"); len(vals) > 0 {
		const prefix = "Bearer "
		authz := vals[0]
		if strings.HasPrefix(authz, prefix) {
			token := strings.TrimPrefix(authz, prefix)
			return validateJWT(token, cfg.JWTSecret)
		}
	}

	if vals := md.Get("x-internal-auth"); len(vals) > 0 {
		if cfg.InternalAuthToken == "" {
			return "", false
		}
		if subtle.ConstantTimeCompare([]byte(vals[0]), []byte(cfg.InternalAuthToken)) == 1 {
			if userVals := md.Get("x-user-id"); len(userVals) > 0 && userVals[0] != "" {
				return userVals[0], true
			}
			return "", true
		}
		return validateJWT(vals[0], cfg.InternalAuthToken)
	}

	return "", false
}

// GRPCAuthUnaryInterceptor returns a unary interceptor that validates auth and
// attaches the user ID to the request context.
func GRPCAuthUnaryInterceptor(cfg *config.Config) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		userID, ok := grpcAuth(ctx, cfg)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "unauthenticated")
		}
		ctx = withUserID(ctx, userID)
		return handler(ctx, req)
	}
}

// GRPCAuthStreamInterceptor returns a stream interceptor that validates auth and
// attaches the user ID to the stream context.
func GRPCAuthStreamInterceptor(cfg *config.Config) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		userID, ok := grpcAuth(stream.Context(), cfg)
		if !ok {
			return status.Error(codes.Unauthenticated, "unauthenticated")
		}
		stream = &contextServerStream{ServerStream: stream, ctx: withUserID(stream.Context(), userID)}
		return handler(srv, stream)
	}
}

type contextServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *contextServerStream) Context() context.Context {
	return s.ctx
}
