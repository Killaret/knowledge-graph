package user

import "errors"

var (
	ErrLoginRequired        = errors.New("login is required")
	ErrEmailRequired        = errors.New("email is required")
	ErrPasswordHashRequired = errors.New("password hash is required")
	ErrAPIKeyHashRequired   = errors.New("API key hash is required")
	ErrAPIKeyNameRequired   = errors.New("API key name is required")
	ErrRoleNotFound         = errors.New("role not found")
)
