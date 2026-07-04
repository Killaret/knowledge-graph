package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository is the repository for working with users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create saves a new user
func (r *UserRepository) Create(ctx context.Context, user *UserModel) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID looks up a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*UserModel, error) {
	var user UserModel
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByLogin looks up a user by login (we use Login instead of Email)
func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*UserModel, error) {
	var user UserModel
	err := r.db.WithContext(ctx).First(&user, "login = ?", login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates user data
func (r *UserRepository) Update(ctx context.Context, user *UserModel) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft-deletes a user (if DeletedAt exists) or removes it entirely
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id).Error
}

// Exists checks whether a user exists
func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
