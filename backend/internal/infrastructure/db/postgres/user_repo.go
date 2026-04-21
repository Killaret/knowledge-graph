package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository — репозиторий для работы с пользователями
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый репозиторий
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create сохраняет нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *UserModel) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID ищет пользователя по ID
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

// FindByLogin ищет пользователя по логину (используем Login вместо Email)
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

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *UserModel) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete мягко удаляет пользователя (если есть DeletedAt) или полностью
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id).Error
}

// Exists проверяет существование пользователя
func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
