# Cursor Rule: knowledge-graph-security

Security requirements, authentication patterns, and threat mitigations for the
Knowledge Graph stack.

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

---

## Never Commit Secrets

Files that must NEVER be committed:
```
.env
.env.local
.env.*.local
*.pem, *.key, *.p12
huggingface_cache/   (large binary, not secret but gitignored)
backups/
```

Pre-commit check:
```bash
# Scan for secrets before push
git diff --staged | grep -iE "(password|secret|token|api_key)\s*=" && echo "SECRETS DETECTED" || echo "OK"

# Verify .gitignore covers .env
grep "^\.env" .gitignore
```

Required `.env` secrets (never hard-code defaults in source):
```
JWT_SECRET=
POSTGRES_PASSWORD=
BACKUP_YANDEX_TOKEN=
GRAPH_SERVICE_INTERNAL_TOKEN=
```

---

## JWT Authentication Middleware

```go
// backend/internal/auth/jwt.go
manager := auth.NewJWTManager(
    cfg.JWTSecret,         // from env — never hard-coded
    15*time.Minute,        // access token TTL
    7*24*time.Hour,        // refresh token TTL
)

// Validate token — always specify expected type ("access" or "refresh")
claims, err := manager.ValidateToken(tokenString, "access")
if err != nil {
    c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid token"})
    c.Abort()
    return
}

// Check blacklist (logout / token revocation)
blacklisted, err := tokenStore.IsTokenBlacklisted(ctx, tokenString)
if blacklisted {
    c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "token revoked"})
    c.Abort()
    return
}
```

Middleware wired in `backend/cmd/server/router.go`:
```go
r.Use(middleware.JWTAuth(jwtConfig))
```

---

## Redis Token Store Security

```go
// backend/internal/auth/redis_store.go
// Tokens are NEVER stored raw — always SHA-256 hashed
func hashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}

// Redis key scheme (no PII in key names)
auth:blacklist:{sha256(token)}    ← TTL = remaining token lifetime
auth:refresh:{sha256(token)}      ← TTL = refresh token lifetime
auth:pkce:{state}                 ← TTL = OAuth session (short)
auth:password_reset:{sha256(tok)} ← TTL = 15 minutes
```

---

## CORS Configuration

```go
// backend/cmd/server/middleware.go
func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        if origin == "" { origin = "*" }
        c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
        c.Writer.Header().Set("Access-Control-Allow-Methods",
            "GET, POST, PUT, DELETE, OPTIONS, PATCH")
        c.Writer.Header().Set("Access-Control-Allow-Headers",
            "Content-Type, Authorization, Accept, Origin, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

In production, replace `origin` echo with an explicit allowlist:
```go
allowed := map[string]bool{"https://yourdomain.com": true}
if !allowed[origin] { origin = "" }
```

---

## Rate Limiting for Write Operations

All `POST`, `PUT`, `DELETE` routes use `writeLimiter`:

```go
// backend/cmd/server/router.go
v1.POST("/notes", writeLimiter, noteHandler.Create)
v1.PUT("/notes/:id", writeLimiter, noteHandler.Update)
v1.DELETE("/notes/:id", writeLimiter, noteHandler.Delete)
v1.POST("/links", writeLimiter, linkHandler.Create)

// backend/cmd/server/middleware.go — conditional by config
func newWriteLimiter(cfg *config.Config) gin.HandlerFunc {
    if !cfg.ServerRateLimitEnabled {
        return func(c *gin.Context) { c.Next() }
    }
    endpointLimits := map[string]int{
        "/notes":     cfg.ServerRateLimitEndpoints["notes_create"],
        "/links":     cfg.ServerRateLimitEndpoints["links_create"],
        "/notes/:id": cfg.ServerRateLimitEndpoints["notes_update"],
    }
    return middleware.RateLimitByEndpoint(endpointLimits,
        cfg.ServerRateLimitRequests, rateWindow)
}
```

---

## Input Validation

Always use `go-playground/validator/v10` tags on request DTOs:

```go
type CreateNoteRequest struct {
    Title    string `json:"title"    validate:"required,min=1,max=200"`
    Content  string `json:"content"  validate:"max=10000"`
    Type     string `json:"type"     validate:"omitempty,oneof=star highlight link"`
}

// In handler — ShouldBindJSON runs validator automatically with Gin
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
    return
}
```

Domain Value Objects provide a second validation layer:
```go
title, err := note.NewTitle(req.Title)  // enforces domain constraints too
```

---

## Yandex OAuth Token Handling

```go
// backend/internal/interfaces/api/handlers/auth/handler.go
// Flow: YandexLogin → redirect to Yandex → callback with code+state

// State parameter is stored in Redis (anti-CSRF)
func (h *Handler) YandexLogin(c *gin.Context) {
    pkce, _ := auth.GeneratePKCE(32)
    state := uuid.New().String()
    h.tokenStore.StorePKCE(ctx, state, pkce, 10*time.Minute)
    // Redirect to Yandex OAuth URL with state + code_challenge
}

func (h *Handler) YandexCallback(c *gin.Context) {
    state := c.Query("state")
    pkce, err := h.tokenStore.GetPKCE(ctx, state)  // one-time retrieval
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid state"})
        return
    }
    // Exchange code for Yandex token, create/load user, issue JWT pair
}
```

`BACKUP_YANDEX_TOKEN` for Yandex.Disk backup is a separate service account
token — never reused for OAuth user authentication.

---

## Environment Variable Patterns

```go
// backend/internal/config/config.go
type Config struct {
    JWTSecret     string  // read from JWT_SECRET env — panic if empty
    DatabaseURL   string  // DATABASE_URL
    RedisURL      string  // REDIS_URL
    SkipAuth      bool    // SKIP_AUTH — false in production ALWAYS
}

// Load with godotenv + os.Getenv
if err := godotenv.Load(); err != nil {
    log.Println("No .env file, reading from environment")
}
if cfg.JWTSecret == "" {
    log.Fatal("JWT_SECRET must be set")
}
```

`SKIP_AUTH=true` is only for automated testing. Verify it is `false` in all
production deployments.

---

## Security Anti-Patterns

```go
// ❌ Hard-coded secret
manager := auth.NewJWTManager("mysecret123", ...)

// ❌ Skipping token type check
claims, _ := manager.ValidateToken(tokenString, "")  // accepts any token type

// ❌ Not checking blacklist (allows logged-out tokens)
claims, err := manager.ValidateToken(tokenString, "access")
// Missing: tokenStore.IsTokenBlacklisted()

// ❌ Logging token or password
log.Printf("User logged in with token: %s", tokenString)

// ❌ SKIP_AUTH=true in docker-compose.yml committed to repo
environment:
  SKIP_AUTH: "true"  // NEVER commit this

// ❌ Rate limit disabled globally
cfg.ServerRateLimitEnabled = false  // OK for dev, NEVER in production
```
