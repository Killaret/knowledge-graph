// Package middleware provides HTTP middleware for the API
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// LogEntry represents a structured log entry for API requests
type LogEntry struct {
	Timestamp    string                 `json:"timestamp"`
	UserID       string                 `json:"user_id,omitempty"`
	Login        string                 `json:"login,omitempty"`
	Role         string                 `json:"role,omitempty"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	ClientIP     string                 `json:"client_ip"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	StatusCode   int                    `json:"status_code"`
	Latency      string                 `json:"latency"`
	DBEntity     string                 `json:"db_entity,omitempty"`
	DBOperation  string                 `json:"db_operation,omitempty"`
	RequestData  map[string]interface{} `json:"request_data,omitempty"`
	ResponseData map[string]interface{} `json:"response_data,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// LoggingMiddleware logs API requests with token data and minimal necessary information
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Extract token data if available
		var userID, login, role string
		if claims, exists := c.Get("token_claims"); exists {
			if tc, ok := claims.(map[string]interface{}); ok {
				if uid, ok := tc["user_id"]; ok {
					if uidStr, ok := uid.(string); ok {
						userID = uidStr
					} else if uidUUID, ok := uid.(uuid.UUID); ok {
						userID = uidUUID.String()
					}
				}
				if l, ok := tc["login"]; ok {
					login = l.(string)
				}
				if r, ok := tc["role"]; ok {
					role = r.(string)
				}
			}
		}

		// Read request body for logging (limited size)
		var requestData map[string]interface{}
		if c.Request.Body != nil && c.Request.Method != "GET" {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil && len(bodyBytes) > 0 && len(bodyBytes) < 10000 { // Limit to 10KB
				// Restore body for next handlers
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				_ = json.Unmarshal(bodyBytes, &requestData)
			}
		}

		// Create response writer wrapper to capture response
		w := &responseWriter{ResponseWriter: c.Writer}
		c.Writer = w

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Extract DB entity and operation from context if set by handlers
		dbEntity, _ := c.Get("db_entity")
		dbOperation, _ := c.Get("db_operation")

		// Create log entry
		entry := LogEntry{
			Timestamp:   time.Now().UTC().Format(time.RFC3339),
			UserID:      userID,
			Login:       login,
			Role:        role,
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			ClientIP:    c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			StatusCode:  w.status,
			Latency:     latency.String(),
			RequestData: requestData,
		}

		if dbEntity != nil {
			if entityStr, ok := dbEntity.(string); ok {
				entry.DBEntity = entityStr
			}
		}
		if dbOperation != nil {
			if opStr, ok := dbOperation.(string); ok {
				entry.DBOperation = opStr
			}
		}

		// Log error if present
		if len(c.Errors) > 0 {
			entry.Error = c.Errors.String()
		}

		// Output structured log
		logJSON, _ := json.Marshal(entry)
		log.Printf("[API_LOG] %s", string(logJSON))
	}
}

// responseWriter wraps gin.ResponseWriter to capture status code
type responseWriter struct {
	gin.ResponseWriter
	status int
}

// WriteHeader captures the status code
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// SetDBEntity sets the database entity being accessed in the context
func SetDBEntity(c *gin.Context, entity string) {
	c.Set("db_entity", entity)
}

// SetDBOperation sets the database operation being performed in the context
func SetDBOperation(c *gin.Context, operation string) {
	c.Set("db_operation", operation)
}
