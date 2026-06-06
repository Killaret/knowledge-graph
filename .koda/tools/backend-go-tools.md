# Инструменты Backend Go Агента

## 🎯 Основные задачи

1. Разработка API endpoints
2. Работа с БД (PostgreSQL, MongoDB)
3. gRPC сервисы
4. Кэширование (Redis)
5. Фоновые задачи (Queues)
6. Тестирование

---

## 🛠️ Разработка

### 1. API Development

#### REST API Patterns
```go
// Handler structure
type NoteHandler struct {
    service *NoteService
    logger  *Logger
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

// Routing
func SetupRoutes(r *gin.Engine, handler *NoteHandler) {
    notes := r.Group("/api/v1/notes")
    notes.Use(auth.Middleware())
    {
        notes.POST("", handler.CreateNote)
        notes.GET("/:id", handler.GetNote)
        notes.PUT("/:id", handler.UpdateNote)
        notes.DELETE("/:id", handler.DeleteNote)
    }
}
```

#### gRPC Service
```go
// Proto definition
syntax = "proto3";
package graph;

service GraphService {
    rpc GetNoteGraph (NoteRequest) returns (GraphResponse);
    rpc GetFullGraph (FullGraphRequest) returns (FullGraphResponse);
}

// Server implementation
type GraphServer struct {
    pb.UnimplementedGraphServiceServer
    repo *GraphRepository
}

func (s *GraphServer) GetNoteGraph(ctx context.Context, req *pb.NoteRequest) (*pb.GraphResponse, error) {
    graph, err := s.repo.GetGraphByNoteID(ctx, req.GetNoteId(), int(req.GetDepth()))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    
    return &pb.GraphResponse{
        Nodes: graph.Nodes,
        Links: graph.Links,
    }, nil
}
```

### 2. Database Operations

#### Repository Pattern
```go
// Interface
type NoteRepository interface {
    Create(ctx context.Context, note *Note) error
    GetByID(ctx context.Context, id string) (*Note, error)
    Update(ctx context.Context, note *Note) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]*Note, error)
}

// PostgreSQL implementation
type PostgresNoteRepo struct {
    db *sql.DB
}

func (r *PostgresNoteRepo) Create(ctx context.Context, note *Note) error {
    query := `
        INSERT INTO notes (id, title, content, type, metadata, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
    _, err := r.db.ExecContext(ctx, query,
        note.ID, note.Title, note.Content, note.Type,
        note.Metadata, note.CreatedAt, note.UpdatedAt)
    return err
}

// MongoDB implementation
type MongoNoteRepo struct {
    collection *mongo.Collection
}

func (r *MongoNoteRepo) Create(ctx context.Context, note *Note) error {
    _, err := r.collection.InsertOne(ctx, note)
    return err
}
```

#### Connection Pooling
```go
// PostgreSQL pool
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

// MongoDB client
func NewMongoClient(uri string) (*mongo.Client, error) {
    opts := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(context.TODO(), opts)
    if err != nil {
        return nil, err
    }
    
    if err := client.Ping(context.TODO(), nil); err != nil {
        return nil, err
    }
    
    return client, nil
}
```

### 3. Caching (Redis)

#### Cache Service
```go
type CacheService struct {
    redis *redis.Client
    ttl   time.Duration
}

func (s *CacheService) Get(ctx context.Context, key string) ([]byte, error) {
    data, err := s.redis.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, ErrNotFound
    }
    return data, err
}

func (s *CacheService) Set(ctx context.Context, key string, value []byte) error {
    return s.redis.Set(ctx, key, value, s.ttl).Err()
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
    return s.redis.Del(ctx, key).Err()
}

// Cache-aside pattern
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
    data := serializeGraph(graph)
    s.cache.Set(ctx, cacheKey, data)
    
    return graph, nil
}
```

### 4. Queue Processing

#### RabbitMQ Integration
```go
type QueueService struct {
    conn    *amqp.Connection
    channel *amqp.Channel
}

func (s *QueueService) Publish(ctx context.Context, queue string, msg []byte) error {
    return s.channel.Publish(
        "",
        queue,
        false,
        false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        msg,
        },
    )
}

func (s *QueueService) Consume(queue string, handler func([]byte) error) error {
    msgs, err := s.channel.Consume(
        queue,
        "",
        false,
        false,
        false,
        false,
        nil,
    )
    if err != nil {
        return err
    }
    
    for msg := range msgs {
        if err := handler(msg.Body); err != nil {
            msg.Nack(false, true)
        } else {
            msg.Ack(false)
        }
    }
    
    return nil
}
```

#### Worker Pool
```go
type Worker struct {
    id      int
    jobs    chan Job
    results chan Result
}

func (w *Worker) Start() {
    go func() {
        for job := range w.jobs {
            result := w.process(job)
            w.results <- result
        }
    }()
}

type Pool struct {
    workers []*Worker
    jobs    chan Job
}

func NewPool(numWorkers int, jobQueueSize int) *Pool {
    pool := &Pool{
        jobs: make(chan Job, jobQueueSize),
    }
    
    for i := 0; i < numWorkers; i++ {
        worker := &Worker{
            id:   i,
            jobs: pool.jobs,
        }
        worker.Start()
        pool.workers = append(pool.workers, worker)
    }
    
    return pool
}
```

### 5. Authentication & Authorization

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

// Permission check
func RequirePermission(permission string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetString("userID")
        hasPermission, err := checkUserPermission(userID, permission)
        if err != nil || !hasPermission {
            c.JSON(403, ErrorResponse{Code: "FORBIDDEN"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

## 🧪 Тестирование

### Unit Tests
```go
func TestNoteService_CreateNote(t *testing.T) {
    mockRepo := new(MockNoteRepository)
    service := NewNoteService(mockRepo)
    
    req := CreateNoteRequest{
        Title:   "Test Note",
        Content: "Test Content",
        Type:    "star",
    }
    
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*Note")).Return(nil)
    
    note, err := service.CreateNote(context.Background(), req)
    
    assert.NoError(t, err)
    assert.Equal(t, "Test Note", note.Title)
    mockRepo.AssertCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestNoteHandler_CreateNote(t *testing.T) {
    router := gin.Default()
    handler := NewNoteHandler(NewNoteService(nil))
    SetupRoutes(router, handler)
    
    reqBody := `{"title":"Test","content":"Content","type":"star"}`
    req, _ := http.NewRequest("POST", "/api/v1/notes", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
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
    
    // Create note
    createReq := `{"title":"Integration Test","content":"Test","type":"star"}`
    req, _ := http.NewRequest("POST", "/api/v1/notes", strings.NewReader(createReq))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 201, w.Code)
    
    // Get note
    var created Note
    json.Unmarshal(w.Body.Bytes(), &created)
    
    getReq, _ := http.NewRequest("GET", "/api/v1/notes/"+created.ID, nil)
    w = httptest.NewRecorder()
    router.ServeHTTP(w, getReq)
    
    assert.Equal(t, 200, w.Code)
}
```

### Mocks
```go
//go:generate mockery --name NoteRepository --output mocks
type NoteRepository interface {
    Create(ctx context.Context, note *Note) error
    GetByID(ctx context.Context, id string) (*Note, error)
    Update(ctx context.Context, note *Note) error
    Delete(ctx context.Context, id string) error
}

// Mock implementation
type MockNoteRepository struct {
    mock.Mock
}

func (m *MockNoteRepository) Create(ctx context.Context, note *Note) error {
    args := m.Called(ctx, note)
    return args.Error(0)
}

func (m *MockNoteRepository) GetByID(ctx context.Context, id string) (*Note, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Note), args.Error(1)
}
```

---

## 📊 Метрики

### Prometheus Integration
```go
var (
    httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    }, []string{"method", "endpoint", "status"})
    
    httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "Duration of HTTP requests",
        Buckets: prometheus.DefBuckets,
    }, []string{"method", "endpoint"})
    
    dbConnections = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "db_connections_active",
        Help: "Number of active database connections",
    })
)

// Middleware
func PrometheusMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        c.Next()
        
        httpRequests.WithLabelValues(
            c.Request.Method,
            path,
            strconv.Itoa(c.Writer.Status()),
        ).Inc()
        
        httpDuration.WithLabelValues(
            c.Request.Method,
            path,
        ).Observe(time.Since(start).Seconds())
    }
}
```

---

## 🛡️ Best Practices

### Error Handling
```go
// Custom error types
type AppError struct {
    Code    string
    Message string
    Details map[string]interface{}
    Err     error
}

func (e *AppError) Error() string {
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error wrapping
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

// Usage
note, err := repo.GetByID(ctx, id)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, wrapError(err, "NOT_FOUND", "Note not found")
    }
    return nil, wrapError(err, "INTERNAL_ERROR", "Failed to get note")
}
```

### Context Usage
```go
// Always use context
func GetUser(ctx context.Context, id string) (*User, error) {
    // Check timeout
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }
    
    // Pass context to DB
    row := db.QueryRowContext(ctx, "SELECT * FROM users WHERE id = $1", id)
    
    // ...
}

// Context with timeout
func ProcessWithTimeout(ctx context.Context, data string) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    return processData(ctx, data)
}
```

### Logging
```go
// Structured logging
type Logger struct {
    log *zap.SugaredLogger
}

func (l *Logger) Info(msg string, fields ...interface{}) {
    l.log.Infow(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...interface{}) {
    l.log.Errorw(msg, fields...)
}

// Usage in handlers
logger.Info("Request processed",
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
    "userID", userID,
    "duration", time.Since(start))
```

---

## 🔧 Команды

### Запуск тестов
```powershell
# Все тесты
go test -race -cover ./...

# Конкретный пакет
go test -v ./internal/application/graph

# С coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Генерация mocks
```powershell
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate
go generate ./...
```

### Linting
```powershell
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.54.0

# Run
golangci-lint run ./...
golangci-lint run --fix
```

### Build
```powershell
# Development
go build -o server ./cmd/server

# Production
go build -ldflags="-w -s" -o server ./cmd/server

# Cross-platform
GOOS=linux GOARCH=amd64 go build -o server ./cmd/server
```
