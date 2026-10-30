package postgres

import (
	"context"
	"errors"

	"knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) FindByName(ctx context.Context, name string) (uuid.UUID, error) {
	var role UserRoleModel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, user.ErrRoleNotFound
	}
	if err != nil {
		return uuid.Nil, err
	}
	return role.ID, nil
}

func (r *RoleRepository) FindByID(ctx context.Context, id uuid.UUID) (string, error) {
	var role UserRoleModel
	err := r.db.WithContext(ctx).First(&role, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", user.ErrRoleNotFound
	}
	if err != nil {
		return "", err
	}
	return role.Name, nil
}
