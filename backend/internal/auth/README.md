# Authentication and Authorization System

This package provides a complete authentication and authorization system for the Knowledge Graph application.

## Features

### 1. Password Authentication with Argon2id
- Secure password hashing using Argon2id
- Configurable parameters (time, memory, threads)
- Password policy enforcement (length, complexity)

### 2. JWT Token Management
- Access tokens (short-lived, 15 minutes by default)
- Refresh tokens (long-lived, 7 days by default)
- Token rotation on refresh
- Token blacklisting for logout

### 3. Yandex OAuth with PKCE
- OAuth 2.0 flow support
- PKCE (Proof Key for Code Exchange) for enhanced security
- Configurable through environment variables or config file

### 4. API Key Authentication
- API keys for service-to-service authentication
- Scoping support
- Usage tracking

### 5. Role-Based Access Control (RBAC)
- Predefined roles: admin, user, guest
- Fine-grained permissions per role
- Permission caching in Redis

### 6. Note Sharing
- Direct user-to-user sharing
- Public share links with tokens
- Permission levels: read, write
- Expiration and usage limits

### 7. Creator Tracking
- Notes and links track their creator
- Owner-based access control
- Soft deletion support

## Configuration

Add the following to `knowledge-graph.config.json`:

```json
{
  "backend": {
    "auth": {
      "jwt_secret": "your-secret-key",
      "jwt_access_ttl_seconds": 900,
      "jwt_refresh_ttl_seconds": 604800,
      "argon2_time": 3,
      "argon2_memory": 65536,
      "argon2_threads": 4,
      "api_key_enabled": true,
      "yandex_client_id": "",
      "yandex_client_secret": "",
      "pkce_enabled": true,
      "pkce_code_challenge_length": 128,
      "password_reset_ttl_seconds": 900,
      "password_policy_min_length": 10,
      "password_policy_require_upper": true,
      "password_policy_require_lower": true,
      "password_policy_require_digit": true,
      "password_policy_require_special": true
    }
  }
}
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh token pair
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password
- `GET /api/v1/auth/yandex/login` - Yandex OAuth login
- `GET /api/v1/auth/yandex/callback` - Yandex OAuth callback

### User Management
- `GET /api/v1/users/me` - Get current user
- `PUT /api/v1/users/me` - Update current user
- `DELETE /api/v1/users/me` - Delete current user (soft)
- `GET /api/v1/users/me/api-keys` - List API keys
- `POST /api/v1/users/me/api-keys` - Create API key
- `DELETE /api/v1/users/me/api-keys/:id` - Revoke API key

### Note Sharing
- `POST /api/v1/notes/:id/share` - Share note with user
- `POST /api/v1/notes/:id/share-link` - Create share link
- `GET /api/v1/notes/:id/shares` - List note shares
- `DELETE /api/v1/share-links/:id` - Revoke share link
- `GET /api/v1/share/:token` - Access shared note

## Middleware Usage

### JWT Authentication
```go
jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
tokenStore := auth.NewRedisTokenStore(redisClient)

jwtConfig := middleware.DefaultJWTConfig(jwtManager, tokenStore)
r.Use(middleware.JWTAuth(jwtConfig))
```

### API Key Authentication
```go
apiKeyConfig := middleware.DefaultAPIKeyConfig(db, cfg.APIKeyEnabled)
r.Use(middleware.APIKey(apiKeyConfig))
```

### Permission Check
```go
permConfig := middleware.DefaultPermissionConfig(db, tokenStore)
r.GET("/admin/data", middleware.Can(permConfig, "data", "read"), handler)
```

### Require Role
```go
r.GET("/admin/users", middleware.RequireAdmin(), handler)
```

## Database Schema

The migration `016_add_auth_and_sharing.sql` creates:

- **users** - Extended with `deleted_at`, `email`, `role_id`
- **user_roles** - Role definitions (admin, user, guest)
- **role_permissions** - Permission assignments
- **notes** - Extended with `creator_id`
- **links** - Extended with `creator_id`
- **note_shares** - Direct note sharing records
- **share_links** - Public share links
- **api_keys** - API key storage
- **refresh_tokens** - Refresh token tracking
- **audit_log** - Security event logging

## Security Considerations

1. **Password Security**
   - Argon2id with configurable parameters
   - Minimum 10 characters with complexity requirements
   - Secure hash storage (never store plaintext)

2. **Token Security**
   - Short-lived access tokens
   - Token rotation on refresh
   - Blacklisting for immediate revocation
   - HTTPS-only transmission

3. **PKCE for OAuth**
   - Prevents authorization code interception
   - Plain method for simplicity (can be upgraded to S256)

4. **Rate Limiting**
   - Apply rate limiting to auth endpoints
   - Separate limits for login attempts

## Testing

Run tests:
```bash
cd backend
go test ./internal/auth/...
```

## Future Enhancements

1. Email service integration for password reset
2. MFA (Multi-Factor Authentication)
3. OAuth providers (Google, GitHub, etc.)
4. Session management UI
5. Advanced audit log analytics
