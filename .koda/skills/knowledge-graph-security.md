# knowledge-graph-security

**Version:** 1.0  
**Purpose:** Security hardening and audit  
**Status:** Active  
**Priority:** 🟡 High

---

## Overview

`knowledge-graph-security` specializes in identifying and resolving security vulnerabilities across the Knowledge Graph application.

**Key Areas:**
- Authentication/Authorization review (JWT, API keys, OAuth)
- Input validation and sanitization
- SQL injection prevention
- XSS prevention
- CORS configuration
- Rate limiting implementation
- Secrets management
- Security headers
- Dependency vulnerability scanning

---

## Security Patterns & Best Practices

### 1. Authentication & Authorization

#### JWT Security

**Target:** Secure token-based authentication

**Techniques:**
```go
// Secure JWT creation
func GenerateJWT(user *User) (string, error) {
    claims := &jwt.MapClaims{
        "user_id": user.ID,
        "email": user.Email,
        "exp": time.Now().Add(15 * time.Minute).Unix(), // Short expiry
        "iat": time.Now().Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    
    // Use strong secret
    secret := os.Getenv("JWT_SECRET")
    if len(secret) < 32 {
        return "", errors.New("JWT_SECRET too weak")
    }
    
    return token.SignedString([]byte(secret))
}

// JWT validation middleware
func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.JSON(401, gin.H{"error": "Missing token"})
            c.Abort()
            return
        }
        
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(os.Getenv("JWT_SECRET")), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }
        
        c.Set("user", token.Claims)
        c.Next()
    }
}
```

**Best Practices:**
- ✅ Use short-lived tokens (15-30 minutes)
- ✅ Implement refresh tokens
- ✅ Store secrets in environment variables
- ✅ Use HS256 or RS256 signing
- ✅ Validate token on every request
- ❌ Don't store sensitive data in JWT payload

#### API Key Security

```go
// API key generation
func GenerateAPIKey() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return base64.URLEncoding.EncodeToString(bytes)
}

// API key middleware
func APIKeyMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            c.JSON(401, gin.H{"error": "Missing API key"})
            c.Abort()
            return
        }
        
        // Validate against database
        if !isValidAPIKey(apiKey) {
            c.JSON(401, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

### 2. Input Validation & Sanitization

#### SQL Injection Prevention

**Target:** Zero SQL injection vulnerabilities

**Techniques:**
```go
// ❌ BAD: String concatenation
query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", userInput)
db.Exec(query)

// ✅ GOOD: Parameterized queries
db.Exec("SELECT * FROM users WHERE id = ?", userID)

// ✅ GOOD: GORM with parameterized queries
var user User
db.Where("id = ?", userID).First(&user)

// ✅ GOOD: Whitelist validation
allowedSortFields := []string{"created_at", "updated_at", "title"}
if !contains(allowedSortFields, sortBy) {
    sortBy = "created_at" // Default
}
db.Order(sortBy).Find(&notes)
```

#### XSS Prevention

**Target:** Zero XSS vulnerabilities

**Techniques:**
```go
// Sanitize user input
func SanitizeHTML(input string) string {
    // Remove script tags
    re := regexp.MustCompile(`<script.*?>.*?</script>`)
    input = re.ReplaceAllString(input, "")
    
    // Remove event handlers
    re = regexp.MustCompile(`on\w+\s*=\s*["'][^"']*["']`)
    input = re.ReplaceAllString(input, "")
    
    // Encode special characters
    return html.EscapeString(input)
}

// Content Security Policy header
func CSPMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Content-Security-Policy", 
            "default-src 'self'; "+
            "script-src 'self'; "+
            "style-src 'self' 'unsafe-inline'; "+
            "img-src 'self' data: https:; "+
            "frame-ancestors 'none';")
        c.Next()
    }
}
```

#### Input Validation Middleware

```go
type InputValidator struct{}

func (v *InputValidator) ValidateNoteCreate(c *gin.Context) error {
    var input struct {
        Title   string `json:"title" binding:"required,min=1,max=200"`
        Content string `json:"content" binding:"max=10000"`
        Tags    []string `json:"tags" binding:"dive,max=50"`
    }
    
    if err := c.ShouldBindJSON(&input); err != nil {
        return err
    }
    
    // Additional validation
    if len(input.Title) == 0 {
        return errors.New("Title is required")
    }
    
    if len(input.Content) > 10000 {
        return errors.New("Content too long")
    }
    
    // Store validated input
    c.Set("validated_input", input)
    return nil
}
```

---

### 3. Rate Limiting & DDoS Protection

#### Rate Limiting

**Target:** Prevent abuse and DDoS

**Techniques:**
```go
// Redis-based rate limiting
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        ip := c.ClientIP()
        key := fmt.Sprintf("ratelimit:%s", ip)
        
        // Get current count
        count, _ := redis.Get(ctx, key).Int()
        
        if count >= maxRequests {
            c.JSON(429, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        
        // Increment counter
        redis.Incr(ctx, key)
        redis.Expire(ctx, key, window)
        
        c.Header("X-RateLimit-Limit", strconv.Itoa(maxRequests))
        c.Header("X-RateLimit-Remaining", strconv.Itoa(maxRequests-count-1))
        
        c.Next()
    }
}

// Usage
r.Use(RateLimitMiddleware(100, 1*time.Minute)) // 100 requests per minute
```

#### Endpoint-Specific Limits

```go
// Sensitive endpoints - stricter limits
r.POST("/api/v1/auth/login", 
    RateLimitMiddleware(5, 1*time.Minute), // 5 attempts per minute
    authHandler.Login)

r.POST("/api/v1/auth/register", 
    RateLimitMiddleware(3, 1*time.Hour), // 3 registrations per hour
    authHandler.Register)

r.POST("/api/v1/notes", 
    RateLimitMiddleware(30, 1*time.Minute), // 30 creates per minute
    tokenMiddleware,
    noteHandler.Create)
```

---

### 4. Security Headers

#### Complete Security Headers Middleware

```go
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Prevent clickjacking
        c.Header("X-Frame-Options", "DENY")
        
        // Prevent MIME type sniffing
        c.Header("X-Content-Type-Options", "nosniff")
        
        // XSS protection
        c.Header("X-XSS-Protection", "1; mode=block")
        
        // Referrer policy
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        
        // Permissions policy
        c.Header("Permissions-Policy", 
            "geolocation=(), microphone=(), camera=()")
        
        // HSTS (only in production)
        if os.Getenv("ENV") == "production" {
            c.Header("Strict-Transport-Security", 
                "max-age=31536000; includeSubDomains")
        }
        
        c.Next()
    }
}
```

---

### 5. Secrets Management

#### Environment Variables

```go
// ❌ BAD: Hardcoded secrets
const jwtSecret = "mysecret123"

// ✅ GOOD: Environment variables
var jwtSecret = os.Getenv("JWT_SECRET")

// Validate required env vars at startup
func validateEnv() {
    required := []string{
        "JWT_SECRET",
        "DATABASE_URL",
        "REDIS_URL",
        "ENCRYPTION_KEY",
    }
    
    for _, env := range required {
        if os.Getenv(env) == "" {
            log.Fatalf("Missing required env var: %s", env)
        }
    }
}
```

#### Encryption Key Management

```go
// Generate secure encryption key
func GenerateEncryptionKey() ([]byte, error) {
    key := make([]byte, 32) // 256-bit
    if _, err := rand.Read(key); err != nil {
        return nil, err
    }
    return key, nil
}

// Store in secure location (AWS Secrets Manager, HashiCorp Vault)
func GetEncryptionKey() ([]byte, error) {
    // Don't store in code or config files
    // Use secrets management service
    return vault.GetSecret("encryption-key")
}
```

---

### 6. Database Security

#### Connection Security

```go
// Secure database connection
func NewPostgresConnection() (*sql.DB, error) {
    // Use SSL
    connStr := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s sslmode=require",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
    )
    
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    
    // Connection pooling
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return db, nil
}
```

#### Query Parameterization

```go
// ❌ BAD: String interpolation
query := fmt.Sprintf("SELECT * FROM notes WHERE user_id = '%s'", userID)

// ✅ GOOD: Parameterized query
query := "SELECT * FROM notes WHERE user_id = $1"
db.QueryRow(query, userID)

// ✅ GOOD: GORM parameterization
var notes []Note
db.Where("user_id = ? AND id = ?", userID, noteID).Find(&notes)
```

---

### 7. CORS Configuration

#### Secure CORS Setup

```go
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Allow specific origins (not *)
        origin := c.Request.Header.Get("Origin")
        allowedOrigins := []string{
            "https://yourdomain.com",
            "https://app.yourdomain.com",
        }
        
        if contains(allowedOrigins, origin) {
            c.Header("Access-Control-Allow-Origin", origin)
        }
        
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        c.Header("Access-Control-Allow-Headers", 
            "Origin, Content-Type, Authorization, X-Requested-With")
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

---

### 8. Dependency Scanning

#### Automated Security Scanning

```bash
# Go dependencies
go get -u github.com/securego/gosec/v2/cmd/gosec
gosec ./...

# npm dependencies
npm audit
npm audit fix

# Python dependencies
pip install safety
safety check
```

#### CI/CD Security Checks

```yaml
# .github/workflows/security.yml
name: Security Scan

on: [push, pull_request]

jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Go Security Scan
        uses: securego/gosec@master
        with:
          args: ./...
      
      - name: npm audit
        run: |
          cd frontend
          npm audit --audit-level=high
      
      - name: Dependency check
        run: |
          cd backend
          go list -m -json all | go-licenses check ./...
```

---

## Security Checklist

### Authentication

- [ ] JWT tokens have short expiry (15-30 min)
- [ ] Refresh tokens implemented
- [ ] Secrets stored in environment variables
- [ ] Strong JWT secret (32+ characters)
- [ ] Token validated on every request
- [ ] Password hashing with bcrypt (cost >= 12)
- [ ] Rate limiting on login endpoint
- [ ] Account lockout after failed attempts

### Authorization

- [ ] Role-based access control (RBAC)
- [ ] Resource-level permissions
- [ ] No IDOR vulnerabilities
- [ ] Admin routes protected
- [ ] API keys scoped to permissions

### Input Validation

- [ ] All inputs validated
- [ ] SQL injection prevented (parameterized queries)
- [ ] XSS prevented (input sanitization, CSP headers)
- [ ] File uploads validated (type, size, scan)
- [ ] Maximum input lengths enforced

### Security Headers

- [ ] HSTS enabled (production)
- [ ] X-Frame-Options: DENY
- [ ] X-Content-Type-Options: nosniff
- [ ] Content-Security-Policy configured
- [ ] Referrer-Policy set
- [ ] Permissions-Policy set

### Data Protection

- [ ] Sensitive data encrypted at rest
- [ ] TLS 1.3 for data in transit
- [ ] API keys rotated regularly
- [ ] Database connections use SSL
- [ ] Secrets not logged

### Monitoring

- [ ] Failed login attempts logged
- [ ] Security events monitored
- [ ] Rate limit violations tracked
- [ ] Unusual patterns detected
- [ ] Alerts configured

---

## Common Vulnerabilities & Fixes

### Issue 1: Weak Password Policy

**Symptom:** Users can set "123456" as password

**Solution:**
```go
func ValidatePassword(password string) error {
    if len(password) < 12 {
        return errors.New("Password must be at least 12 characters")
    }
    
    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
        return errors.New("Password must contain uppercase letter")
    }
    
    if !regexp.MustCompile(`[a-z]`).MatchString(password) {
        return errors.New("Password must contain lowercase letter")
    }
    
    if !regexp.MustCompile(`[0-9]`).MatchString(password) {
        return errors.New("Password must contain number")
    }
    
    if !regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(password) {
        return errors.New("Password must contain special character")
    }
    
    return nil
}
```

---

### Issue 2: Insecure Direct Object Reference (IDOR)

**Symptom:** User can access another user's notes by changing ID

**Solution:**
```go
// ❌ BAD: No ownership check
func GetNote(c *gin.Context) {
    noteID := c.Param("id")
    note := db.GetNote(noteID) // Anyone can access any note
    c.JSON(200, note)
}

// ✅ GOOD: Verify ownership
func GetNote(c *gin.Context) {
    noteID := c.Param("id")
    userID := c.GetString("user_id") // From JWT
    
    note := db.GetNote(noteID)
    if note.UserID != userID {
        c.JSON(403, gin.H{"error": "Access denied"})
        return
    }
    
    c.JSON(200, note)
}
```

---

### Issue 3: Missing Rate Limiting

**Symptom:** API vulnerable to brute force attacks

**Solution:**
```go
// Add rate limiting to sensitive endpoints
r.POST("/api/v1/auth/login",
    RateLimitMiddleware(5, 1*time.Minute),
    authHandler.Login)

r.POST("/api/v1/auth/forgot-password",
    RateLimitMiddleware(3, 1*time.Hour),
    authHandler.ForgotPassword)
```

---

## Security Audit Commands

### Manual Security Testing

```bash
# Test SQL injection
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email": "'\'' OR '\''1'\''='\''1", "password": "test"}'

# Test XSS
curl -X POST http://localhost:8080/api/v1/notes \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "<script>alert(1)</script>"}'

# Check security headers
curl -I http://localhost:8080/health

# Test rate limiting
for i in {1..20}; do
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email": "test@test.com", "password": "test"}'
done
```

### Automated Scanning

```bash
# OWASP ZAP scan
zap-baseline.py -t http://localhost:8080

# Nmap security scan
nmap -sV --script vuln localhost -p 8080

# Go security scanner
gosec ./...

# npm audit
npm audit --audit-level=high
```

---

## References

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [Go Security Best Practices](https://github.com/securego/gosec)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)

---

**Last Updated:** 2026-05-22  
**Maintainer:** knowledge-graph-docs-maintenance  
**Version:** 1.0
