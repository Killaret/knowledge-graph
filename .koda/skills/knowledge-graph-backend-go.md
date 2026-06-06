# knowledge-graph-backend-go

**Version:** 1.0  
**Purpose:** Backend Go development - API, database, microservices  
**Status:** Active  
**Priority:** 🟢 High

---

## Overview

`knowledge-graph-backend-go` specializes in Go backend development for the Knowledge Graph project.

**Key Areas:**
- REST/gRPC API development
- PostgreSQL & MongoDB
- Redis caching
- RabbitMQ queues
- Authentication (JWT, OAuth)
- Unit & integration testing
- Performance optimization
- Structured logging

---

## Development Patterns

### 1. API Development

#### Handler Structure
```go
type NoteHandler struct {
    service *NoteService
    logger  *zap.SugaredLogger
}

func (h *NoteHandler) CreateNote(c *gin.Context) {
    var req CreateNoteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Code: "VALIDATION_ERROR"})
        return
    }
    
    note, err := h.service.CreateNote(c.Request.Context(), req)
    if err != nil {
        h.logger.Error("Failed to create note", "error", err)
        c.JSON(500, ErrorResponse{Code: "INTERNAL_ERROR"})
        return
    }
    
    c.JSON(201, note)
}
```

#### gRPC Service
```go
type GraphServer struct {
    pb.UnimplementedGraphServiceServer
    repo *GraphRepository
}

func (s *GraphServer) GetNoteGraph(ctx context.Context, req *pb.NoteRequest) (*pb.GraphResponse, error) {
    graph, err := s.repo.GetGraphByNoteID(ctx, req.GetNoteId(), int(req.GetDepth()))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &pb.GraphResponse{Nodes: graph.Nodes, Links: graph.Links}, nil
}
```

### 2. Database Patterns

#### Repository Interface
```go
type NoteRepository interface {
    Create(ctx context.Context, note *Note) error
    GetByID(ctx context.Context, id string) (*Note, error)
    Update(ctx context.Context, note *Note) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*Note, error)
}
```

#### PostgreSQL Implementation
```go
type PostgresNoteRepo struct {
    db *sql.DB
}

func NewPostgresPool(dsn string) (*sql.DB, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, err
    }
    
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### 3. Caching (Redis)

#### Cache-Aside Pattern
```go
type CacheService struct {
    redis *redis.Client
    ttl   time.Duration
}

func (s *GraphService) GetGraphWithCache(ctx context.Context, noteID string, depth int) (*Graph, error) {
    cacheKey := fmt.Sprintf("graph:%s:%d", noteID, depth)
    
    // Try cache first
    cached, err := s.cache.Get(ctx, cacheKey)
    if err == nil {
        return deserializeGraph(cached)
    }
    
    // Fetch from DB
    graph, err := s.repo.GetGraph(ctx, noteID, depth)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    s.cache.Set(ctx, cacheKey, serializeGraph(graph))
    return graph, nil
}
```

### 4. Authentication

#### JWT Middleware
```go
func JWTMiddleware(jwtSecret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, ErrorResponse{Code: "UNAUTHORIZED"})
            c.Abort()
            return
        }
        
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, ErrorResponse{Code: "INVALID_TOKEN"})
            c.Abort()
            return
        }
        
        claims := token.Claims.(jwt.MapClaims)
        userID := claims["user_id"].(string)
        c.Set("userID", userID)
        c.Next()
    }
}
```

---

## Testing

### Unit Tests
```go
func TestNoteService_CreateNote(t *testing.T) {
    mockRepo := new(MockNoteRepository)
    service := NewNoteService(mockRepo)
    
    req := CreateNoteRequest{Title: "Test", Content: "Content", Type: "star"}
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*Note")).Return(nil)
    
    note, err := service.CreateNote(context.Background(), req)
    
    assert.NoError(t, err)
    assert.Equal(t, "Test", note.Title)
    mockRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}
```

### Integration Tests
```go
func TestNoteAPI_Integration(t *testing.T) {
    db := setupTestDB(t)
    repo := NewPostgresNoteRepo(db)
    service := NewNoteService(repo)
    handler := NewNoteHandler(service)
    
    router := setupRouter(handler)
    
    createReq := `{"title":"Integration Test","content":"Test","type":"star"}`
    req, _ := http.NewRequest("POST", "/api/v1/notes", strings.NewReader(createReq))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
}
```

---

## Commands

### Run Tests
```bash
# All tests with coverage
go test -race -cover ./...

# Specific package
go test -v ./internal/application/graph

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting
```bash
golangci-lint run ./...
golangci-lint run --fix
```

### Build
```bash
# Development
go build -o server ./cmd/server

# Production
go build -ldflags="-w -s" -o server ./cmd/server

# Cross-platform
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server
```

### Generate Mocks
```bash
go install github.com/vektra/mockery/v2@latest
go generate ./...
```

---

## Best Practices

### Error Handling
```go
type AppError struct {
    Code    string
    Message string
    Details map[string]interface{}
    Err     error
}

func wrapError(err error, code, message string) error {
    if err == nil {
        return nil
    }
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}
```

### Context Usage
```go
// Always use context
func GetUser(ctx context.Context, id string) (*User, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    row := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", id)
    // ...
}
```

### Structured Logging
```go
logger.Info("Request processed",
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
    "userID", userID,
    "duration", time.Since(start))
```

---

## Metrics

### Prometheus Integration
```go
var httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
    Name:    "http_request_duration_seconds",
    Help:    "Duration of HTTP requests",
    Buckets: prometheus.DefBuckets,
}, []string{"method", "endpoint"})

func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        c.Next()
        httpDuration.WithLabelValues(c.Request.Method, path).Observe(
            time.Since(start).Seconds())
    }
}
```

---

**Tools:** `backend-go-tools.md`  
**Coverage Target:** > 60%  
**Response Time Target:** p95 < 500ms