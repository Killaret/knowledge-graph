package postgres

import (
	"time"

	"github.com/google/uuid"
)

// UserModel — пользователь с поддержкой soft delete и ролей
type UserModel struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Login        string         `gorm:"uniqueIndex;not null"`
	Email        string         `gorm:"uniqueIndex"`
	PasswordHash string         `gorm:"not null"`
	RoleID       *uuid.UUID     `gorm:"type:uuid;index"`
	Role         *UserRoleModel `gorm:"foreignKey:RoleID"`
	CreatedAt    time.Time
	DeletedAt    *time.Time `gorm:"index"`
}

func (UserModel) TableName() string { return "users" }

// UserRoleModel — роль пользователя
type UserRoleModel struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Description string
	CreatedAt   time.Time
}

func (UserRoleModel) TableName() string { return "user_roles" }

// RolePermissionModel — разрешения для ролей
type RolePermissionModel struct {
	ID        uuid.UUID     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoleID    uuid.UUID     `gorm:"type:uuid;not null;index"`
	Role      UserRoleModel `gorm:"foreignKey:RoleID"`
	Resource  string        `gorm:"not null"`
	Action    string        `gorm:"not null"`
	CreatedAt time.Time
}

func (RolePermissionModel) TableName() string { return "role_permissions" }

// APIKeyModel — API ключи пользователей
type APIKeyModel struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	User       UserModel `gorm:"foreignKey:UserID"`
	KeyHash    string    `gorm:"not null;index"`
	Name       string    `gorm:"not null"`
	Scopes     []string  `gorm:"type:text[]"`
	CreatedAt  time.Time
	ExpiresAt  *time.Time
	LastUsedAt *time.Time
	IsActive   bool `gorm:"default:true"`
}

func (APIKeyModel) TableName() string { return "api_keys" }

// RefreshTokenModel — refresh токены
type RefreshTokenModel struct {
	ID              uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
	User            UserModel `gorm:"foreignKey:UserID"`
	TokenHash       string    `gorm:"uniqueIndex;not null"`
	DeviceInfo      string
	IPAddress       string
	CreatedAt       time.Time
	ExpiresAt       time.Time `gorm:"not null;index"`
	RevokedAt       *time.Time
	ReplacedByToken *uuid.UUID
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }

// AuditLogModel — лог аудита
type AuditLogModel struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID       *uuid.UUID `gorm:"type:uuid;index"`
	User         *UserModel `gorm:"foreignKey:UserID"`
	Action       string     `gorm:"not null;index"`
	ResourceType string
	ResourceID   *uuid.UUID
	Details      string `gorm:"type:jsonb"`
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time `gorm:"not null;index"`
}

func (AuditLogModel) TableName() string { return "audit_log" }
