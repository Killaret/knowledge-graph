---
name: Security Rules
alwaysApply: true
description: Security policies - secrets management, JWT, CORS, rate limiting, environment variables
---

# Security Rules

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

## NEVER Commit Secrets

These files and patterns MUST NEVER be committed to git:

```gitignore
.env                    # Environment secrets
*.pem                   # Private keys
*.key                   # Private keys
*_secret*               # Any secret files
```

**Check before every commit:**
- No API keys, tokens, or passwords in source code
- No JWT secrets in config files (use env vars)
- No database credentials hardcoded

```go
// ❌ NEVER — hardcoded secrets
const jwtSecret = "my-super-secret-key"
db, _ := gorm.Open(postgres.Open("host=localhost user=admin password=secret123"))

// ✅ ALWAYS — from environment
jwtSecret := os.Getenv("JWT_SECRET")
dbURL := os.Getenv("DATABASE_URL")
```

```typescript
// ❌ NEVER — secrets in frontend code
const API_KEY = "sk-1234567890abcdef";

// ✅ ALWAYS — from environment (server-side only)
const apiKey = import.meta.env.VITE_API_KEY; // Only public keys in VITE_ prefix
```

## JWT Validation Pattern

```go
// JWT middleware — validate on every protected request
func (m *AuthMiddleware) Handle() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract bearer token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header required"})
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")
        if token == authHeader {
            c.AbortWithStatusJSON(401, gin.H{"error": "Bearer token required"})
            return
        }

        // Validate token
        claims, err := m.jwtService.ValidateToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired token"})
            return
        }

        // Set user context
        c.Set("userID", claims.UserID)
        c.Next()
    }
}
```

### JWT Configuration

```json
{
  "auth": {
    "jwt_access_ttl_seconds": 900,       // 15 minutes
    "jwt_refresh_ttl_seconds": 604800,   // 7 days
    "argon2_time": 3,
    "argon2_memory": 65536,
    "argon2_threads": 4
  }
}
```

**Rules:**
- Access tokens: short-lived (15 min)
- Refresh tokens: longer-lived (7 days), stored securely
- Password hashing: Argon2id (never bcrypt for new code)
- Token refresh endpoint must validate refresh token before issuing new access token

## CORS Configuration

```go
func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")
        for _, allowed := range allowedOrigins {
            if origin == allowed {
                c.Header("Access-Control-Allow-Origin", origin)
                break
            }
        }
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
```

**Rules:**
- NEVER use `Access-Control-Allow-Origin: *` with credentials
- Whitelist specific origins only
- Preflight cache: 24 hours max

## Rate Limiting — Write Operations

```go
// Rate limit configuration from knowledge-graph.config.json
type RateLimitConfig struct {
    Enabled       bool `json:"enabled"`
    Requests      int  `json:"requests"`       // 1000 per window
    WindowSeconds int  `json:"window_seconds"` // 60s window
    Endpoints     struct {
        NotesCreate  int `json:"notes_create"`  // 200/min
        LinksCreate  int `json:"links_create"`  // 200/min
        NotesUpdate  int `json:"notes_update"`  // 100/min
        NotesDelete  int `json:"notes_delete"`  // 500/min
    } `json:"endpoints"`
}
```

**Rules:**
- Rate limiting REQUIRED on all write operations (POST, PUT, DELETE)
- Read operations (GET) have higher limits
- Use Redis-based sliding window counter
- Return `429 Too Many Requests` with `Retry-After` header

```go
func RateLimitMiddleware(redis *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := fmt.Sprintf("ratelimit:%s:%s", c.ClientIP(), c.FullPath())
        count, _ := redis.Incr(c.Request.Context(), key).Result()
        if count == 1 {
            redis.Expire(c.Request.Context(), key, window)
        }
        if count > int64(limit) {
            c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
            c.AbortWithStatusJSON(429, gin.H{"error": "Too many requests"})
            return
        }
        c.Next()
    }
}
```

## Environment Variables

All secrets MUST come from environment variables:

| Variable            | Purpose                        | Required |
|---------------------|--------------------------------|----------|
| JWT_SECRET          | JWT signing key                | Yes      |
| POSTGRES_USER       | Database username              | Yes      |
| POSTGRES_PASSWORD   | Database password              | Yes      |
| POSTGRES_DB         | Database name                  | Yes      |
| REDIS_URL           | Redis connection string        | Yes      |
| YANDEX_CLIENT_SECRET| OAuth client secret            | No       |
| SMTP_PASSWORD       | Email service password         | No       |

## Password Policy

```json
{
  "password_policy_min_length": 10,
  "password_policy_require_upper": true,
  "password_policy_require_lower": true,
  "password_policy_require_digit": true,
  "password_policy_require_special": true
}
```

## Security Checklist

- [ ] No secrets in source code or docker-compose files
- [ ] JWT validated on all protected endpoints
- [ ] CORS restricted to known origins
- [ ] Rate limiting on write operations
- [ ] Passwords hashed with Argon2id
- [ ] Environment variables for all sensitive config
- [ ] HTTPS enforced in production (nginx)
- [ ] SQL injection prevented (GORM parameterized queries)
- [ ] Input validation on all endpoints (Pydantic / Value Objects)
- [ ] Error messages don't leak internal details to users
