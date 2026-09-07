# План расширения API тестирования

**Создано:** 29 июля 2026 г.  
**Статус:** ⏳ Запланировано

## Обзор

Документ описывает план расширения API тестирования для покрытия всех endpoint'ов большим количеством кейсов, включая edge cases, валидацию, error handling и security scenarios.

## Текущее состояние

**Существующие тесты:**
- ✅ Basic CRUD operations (Create, Read, Update, Delete)
- ✅ Basic validation (title length, content length)
- ✅ Authentication tests
- ✅ Error responses for invalid data

**Проблемы:**
- ❌ Недостаточное покрытие edge cases
- ❌ Недостаточное тестирование валидации типов
- ❌ Недостаточное тестирование security scenarios
- ❌ Недостаточное тестирование performance scenarios
- ❌ Недостаточное тестирование concurrent operations

## План тестирования по endpoint'ам

### 1. Notes API

#### 1.1 POST /api/v1/notes (Create Note)

**Валидация:**
- ✅ Valid note creation
- ✅ Title validation (empty, too long, valid)
- ✅ Content validation (empty, too long, valid)
- ✅ Type validation (all valid types, invalid types)
- ✅ Metadata validation (valid JSON, invalid JSON)
- ✅ Missing required fields
- ✅ Extra fields in request

**Security:**
- ✅ Authentication required
- ✅ Authorization (user can only create for themselves)
- ✅ XSS in title/content
- ✅ SQL injection attempts
- ✓ CSRF protection (if applicable)

**Edge Cases:**
- ⏳ Create with all supported types (16 types)
- ⏳ Create with null vs empty string for optional fields
- ⏳ Create with Unicode/special characters
- ⏳ Create with extremely long metadata
- ⏳ Create with malformed JSON in metadata
- ⏳ Create with concurrent requests (race conditions)

**Business Logic:**
- ⏳ Default type when not specified (should be "unknown")
- ⏳ IsPublic flag behavior
- ⏳ Creator ID assignment
- ⏳ Timestamps (created_at, updated_at)

#### 1.2 GET /api/v1/notes (List Notes)

**Валидация:**
- ✅ Basic list retrieval
- ⏳ Pagination parameters (limit, offset)
- ⏳ Type filtering (all types)
- ⏳ Search query parameters
- ⏳ Invalid pagination parameters (negative, too large)
- ⏳ Invalid type filters

**Security:**
- ✅ Authentication required
- ✅ User scope (only user's notes)
- ⏳ Access control for public notes
- ⏳ SQL injection in search parameters

**Performance:**
- ⏳ Large dataset response (1000+ notes)
- ⏳ Pagination performance
- ⏳ Complex filter combinations

#### 1.3 GET /api/v1/notes/:id (Get Note)

**Валидация:**
- ✅ Valid note retrieval
- ✅ Invalid UUID format
- ✅ Non-existent note ID
- ⏳ Note belongs to different user (403)
- ⏳ Public note access

**Security:**
- ✅ Authentication required
- ⏳ Authorization checks (user vs public)
- ⏳ IDOR (Insecure Direct Object Reference) prevention

#### 1.4 PUT /api/v1/notes/:id (Update Note)

**Валидация:**
- ✅ Basic update
- ✅ Title validation (empty, too long, valid)
- ✅ Content validation (empty, too long, valid)
- ✅ Type validation (all valid types, invalid types)
- ⏳ Partial update (only title, only content, only type)
- ⏳ Full update (all fields)
- ⏳ No-op update (same values)
- ⏳ Null vs empty string handling

**Security:**
- ✅ Authentication required
- ⏳ Authorization (user can only update their notes)
- ⏳ Prevent type change for system notes (if applicable)

**Business Logic:**
- ⏳ Updated_at timestamp update
- ⏳ Metadata merge vs replace
- ⏳ Type change validation

#### 1.5 DELETE /api/v1/notes/:id (Delete Note)

**Валидация:**
- ✅ Valid deletion
- ✅ Invalid UUID format
- ✅ Non-existent note ID
- ⏳ Note belongs to different user (403)

**Security:**
- ✅ Authentication required
- ⏳ Authorization (user can only delete their notes)
- ⏳ Cascade delete of related data (links, recommendations)

**Business Logic:**
- ⏳ Soft delete vs hard delete
- ⏳ Cleanup of related resources

#### 1.6 POST /api/v1/notes/:id/publish (Publish Note)

**Валидация:**
- ⏳ Valid publish
- ⏳ Invalid UUID format
- ⏳ Non-existent note ID
- ⏳ Already published note
- ⏳ Note belongs to different user

**Security:**
- ⏳ Authentication required
- ⏳ Authorization checks

**Business Logic:**
- ⏳ IsPublic flag update
- ⏳ Updated_at timestamp

#### 1.7 POST /api/v1/notes/:id/unpublish (Unpublish Note)

**Валидация:**
- ⏳ Valid unpublish
- ⏳ Invalid UUID format
- ⏳ Non-existent note ID
- ⏳ Already unpublished note
- ⏳ Note belongs to different user

**Security:**
- ⏳ Authentication required
- ⏳ Authorization checks

**Business Logic:**
- ⏳ IsPublic flag update
- ⏳ Updated_at timestamp

### 2. Links API

#### 2.1 POST /api/v1/links (Create Link)

**Валидация:**
- ✅ Valid link creation
- ✅ Source/target validation (UUID format, existence)
- ✅ Link type validation (all types, invalid types)
- ✅ Weight validation (0-1 range, negative, >1)
- ✅ Metadata validation
- ⏳ Missing required fields
- ⏳ Self-link (source == target)
- ⏳ Duplicate link detection
- ⏳ Reverse link creation

**Security:**
- ✅ Authentication required
- ⏳ Authorization (user can only link their notes)
- ⏳ Link to other user's notes (if public)

**Business Logic:**
- ⏳ Unique constraint (source, target, type)
- ⏳ Weight calculation
- ⏳ SourceType (user vs gamma)

#### 2.2 GET /api/v1/links (List Links)

**Валидация:**
- ⏳ Basic list retrieval
- ⏳ Pagination parameters
- ⏳ Type filtering
- ⏳ Source/target filtering
- ⏳ Weight filtering

**Security:**
- ⏳ Authentication required
- ⏳ User scope filtering

#### 2.3 GET /api/v1/links/:id (Get Link)

**Валидация:**
- ⏳ Valid link retrieval
- ⏳ Invalid UUID format
- ⏳ Non-existent link ID

**Security:**
- ⏳ Authentication required
- ⏳ Authorization checks

#### 2.4 PUT /api/v1/links/:id (Update Link)

**Валидация:**
- ⏳ Basic update
- ⏳ Weight update (0-1 range)
- ⏳ Type update
- ⏳ Partial update
- ⏳ Invalid link type

**Security:**
- ⏳ Authentication required
- ⏳ Authorization checks

#### 2.5 DELETE /api/v1/links/:id (Delete Link)

**Валидация:**
- ⏳ Valid deletion
- ⏳ Invalid UUID format
- ⏳ Non-existent link ID

**Security:**
- ⏳ Authentication required
- ⏳ Authorization checks

### 3. Suggestions API

#### 3.1 GET /api/v1/notes/:id/suggestions (Get Suggestions)

**Валидация:**
- ✅ Valid suggestions request
- ✅ Invalid UUID format
- ✅ Non-existent note ID
- ⏳ Limit parameter validation
- ⏳ Invalid limit parameter

**Performance:**
- ⏳ Empty graph (no suggestions)
- ⏳ Large graph performance
- ⏳ Cache hit vs miss

**Business Logic:**
- ⏳ Score calculation
- ⏳ Ranking algorithm
- ⏳ Component breakdown (alpha, beta, gamma)

### 4. Users API

#### 4.1 GET /api/v1/users/me (Get Current User)

**Валидация:**
- ✅ Valid user retrieval
- ⏳ Missing authentication

**Security:**
- ✅ Authentication required
- ⏳ JWT token validation
- ⏳ Token expiration

#### 4.2 PUT /api/v1/users/me (Update Current User)

**Валидация:**
- ⏳ Email update (valid, invalid, duplicate)
- ⏳ Password update (valid, invalid, weak password)
- ⏳ Profile data update
- ⏳ Partial update

**Security:**
- ⏳ Authentication required
- ⏳ Authorization (can only update self)
- ⏳ Password strength validation

### 5. Authentication API

#### 5.1 POST /api/v1/auth/register (Register)

**Валидация:**
- ✅ Valid registration
- ⏳ Missing required fields
- ⏳ Invalid email format
- ⏳ Weak password
- ⏳ Duplicate username/email
- ⏳ Invalid JSON

**Security:**
- ⏳ Rate limiting
- ⏳ Brute force protection
- ⏳ SQL injection attempts

#### 5.2 POST /api/v1/auth/login (Login)

**Валидация:**
- ✅ Valid login
- ⏳ Missing credentials
- ⏳ Invalid credentials
- ⏳ Non-existent user
- ⏳ Inactive account

**Security:**
- ⏳ Rate limiting
- ⏳ Brute force protection
- ⏳ JWT token generation
- ⏳ Token expiration

#### 5.3 POST /api/v1/auth/refresh (Refresh Token)

**Валидация:**
- ⏳ Valid refresh token
- ⏳ Invalid refresh token
- ⏳ Expired refresh token
- ⏳ Missing refresh token

**Security:**
- ⏳ Token validation
- ⏳ Token rotation
- ⏳ Token revocation

### 6. Graph API

#### 6.1 GET /graph-service/api/v1/graph/all (Get Full Graph)

**Валидация:**
- ⏳ Valid graph request
- ⏳ Authentication required
- ⏳ User scope filtering

**Performance:**
- ⏳ Large graph response
- ⏳ Empty graph response
- ⏳ Cache hit vs miss

#### 6.2 GET /graph-service/api/v1/graph/delta (Get Delta Updates)

**Валидация:**
- ⏳ Valid delta request
- ⏳ Invalid last_hash format
- ⏳ Missing last_hash

**Performance:**
- ⏳ Large delta response
- ⏳ Empty delta response

### 7. Health Check API

#### 7.1 GET /health (Health Check)

**Валидация:**
- ✅ Basic health check
- ⏳ Response format validation
- ⏳ Response time SLA

**Performance:**
- ⏳ Response time under load

## Concurrency Tests

- ⏳ Concurrent note creation
- ⏳ Concurrent note updates
- ⏳ Concurrent link creation
- ⏳ Race conditions in ID generation
- ⏳ Database lock contention

## Performance Tests

- ⏳ API response time baselines
- ⏳ Load testing (100, 1000, 10000 requests)
- ⏳ Database query performance
- ⏳ Cache performance
- ⏳ Memory usage under load

## Security Tests

- ⏳ SQL injection
- ⏳ XSS attacks
- ⏳ CSRF protection
- ⏳ Rate limiting
- ⏳ Authentication bypass attempts
- ⏳ Authorization bypass attempts
- ⏳ Token manipulation
- ⏳ Data leakage tests

## Error Handling Tests

- ⏳ Structured error responses
- ⏳ Error message consistency
- ⏳ HTTP status code correctness
- ⏳ Error response format (JSON structure)
- ⏳ Sensitive data in error messages

## Integration Tests

- ⏳ End-to-end user flows
- ⏳ Multi-service interactions (backend + graph-service + nlp)
- ⏳ Database transaction consistency
- ⏳ Cache invalidation
- ⏳ Event publishing

## Priority Implementation

1. **P0 (Критический):** Core CRUD operations validation and security
2. **P1 (Высокий):** Edge cases and error handling
3. **P2 (Средний):** Performance and concurrency tests
4. **P3 (Низкий):** Advanced security scenarios

## Связанные задачи

- Документация: `docs/TESTING.md` — общие стратегии тестирования
- Документация: `docs/REGRESSION_TEST_PLAN.md` — регрессионное тестирование
- Код: `backend/internal/interfaces/api/` — API endpoint'ы