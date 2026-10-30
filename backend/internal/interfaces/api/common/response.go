package common

import (
	"github.com/gin-gonic/gin"
)

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Data    any    `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError represents a detailed field validation error
type FieldError struct {
	Field    string `json:"field"`
	Reason   string `json:"reason"`
	Message  string `json:"message"`
	Received any    `json:"received,omitempty"`
	Expected string `json:"expected,omitempty"`
}

// Error codes
const (
	ErrCodeValidationError = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeInternalError   = "INTERNAL_ERROR"
	ErrCodeDuplicateLink   = "DUPLICATE_LINK"
	ErrCodeInvalidUUID     = "INVALID_UUID"
	ErrCodeInvalidRequest  = "INVALID_REQUEST"
)

// Error messages in English
const (
	MsgValidationError   = "Invalid input data"
	MsgNotFound          = "Entity not found"
	MsgConflict          = "Data conflict"
	MsgForbidden         = "Access forbidden"
	MsgUnauthorized      = "Authentication required"
	MsgInternalError     = "Internal server error"
	MsgDuplicateLink     = "Link already exists"
	MsgInvalidUUID       = "Invalid UUID format"
	MsgInvalidRequest    = "Invalid request"
	MsgResourceCreated   = "Resource created successfully"
	MsgResourceUpdated   = "Resource updated successfully"
	MsgResourceDeleted   = "Resource deleted successfully"
	MsgNoteNotFound      = "Note not found"
	MsgLinkNotFound      = "Link not found"
	MsgSourceNotFound    = "Source note not found"
	MsgTargetNotFound    = "Target note not found"
	MsgFailedSaveNote    = "Failed to save note"
	MsgFailedUpdateNote  = "Failed to update note"
	MsgFailedDeleteNote  = "Failed to delete note"
	MsgFailedFetchNote   = "Failed to fetch note"
	MsgFailedSaveLink    = "Failed to save link"
	MsgFailedDeleteLink  = "Failed to delete link"
	MsgFailedFetchLink   = "Failed to fetch link"
	MsgFailedFetchNotes  = "Failed to fetch notes"
	MsgFailedSearchNotes = "Failed to search notes"
	MsgFailedLoadGraph   = "Failed to load graph"
)

// Validation reasons
const (
	ReasonRequired      = "required"
	ReasonInvalidType   = "invalid_type"
	ReasonInvalidFormat = "invalid_format"
	ReasonAlreadyExists = "already_exists"
	ReasonTooLong       = "too_long"
	ReasonTooShort      = "too_short"
	ReasonOutOfRange    = "out_of_range"
	ReasonInvalidValue  = "invalid_value"
)

// JSON sends a success response
func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, SuccessResponse{Data: data})
}

// JSONWithMessage sends a success response with message
func JSONWithMessage(c *gin.Context, status int, data any, msg string) {
	c.JSON(status, SuccessResponse{Data: data, Message: msg})
}

// Error sends error response
func Error(c *gin.Context, status int, code, msg string) {
	c.JSON(status, ErrorResponse{Code: code, Message: msg})
}

// ErrorWithDetails sends error with field details
func ErrorWithDetails(c *gin.Context, status int, code, msg string, details []FieldError) {
	c.JSON(status, ErrorResponse{Code: code, Message: msg, Details: details})
}

// NoContent sends 204
func NoContent(c *gin.Context) {
	c.Status(204)
}

// BadRequest sends 400 with details
func BadRequest(c *gin.Context, details []FieldError) {
	ErrorWithDetails(c, 400, ErrCodeValidationError, MsgValidationError, details)
}

// BadRequestSimple sends 400 with message
func BadRequestSimple(c *gin.Context, msg string) {
	Error(c, 400, ErrCodeInvalidRequest, msg)
}

// NotFound sends 404
func NotFound(c *gin.Context, entity string) {
	m := MsgNotFound
	if entity != "" {
		m = entity + " not found"
	}
	Error(c, 404, ErrCodeNotFound, m)
}

// Conflict sends 409
func Conflict(c *gin.Context, details []FieldError) {
	ErrorWithDetails(c, 409, ErrCodeConflict, MsgConflict, details)
}

// ConflictSimple sends 409 with message
func ConflictSimple(c *gin.Context, msg string) {
	Error(c, 409, ErrCodeConflict, msg)
}

// InternalError sends 500
func InternalError(c *gin.Context) {
	Error(c, 500, ErrCodeInternalError, MsgInternalError)
}

// InternalErrorWithMessage sends 500 with custom message
func InternalErrorWithMessage(c *gin.Context, msg string) {
	Error(c, 500, ErrCodeInternalError, msg)
}

// Forbidden sends 403
func Forbidden(c *gin.Context) {
	Error(c, 403, ErrCodeForbidden, MsgForbidden)
}

// Unauthorized sends 401
func Unauthorized(c *gin.Context) {
	Error(c, 401, ErrCodeUnauthorized, MsgUnauthorized)
}

// NewFieldError creates field error
func NewFieldError(field, reason, msg string) FieldError {
	return FieldError{Field: field, Reason: reason, Message: msg}
}

// NewFieldErrorWithValue creates field error with value
func NewFieldErrorWithValue(field, reason, msg string, received any) FieldError {
	return FieldError{Field: field, Reason: reason, Message: msg, Received: received}
}

// NewFieldErrorFull creates full field error
func NewFieldErrorFull(field, reason, msg string, received any, expected string) FieldError {
	return FieldError{Field: field, Reason: reason, Message: msg, Received: received, Expected: expected}
}
