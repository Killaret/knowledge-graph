package postgres

import (
	"context"
	"errors"
	"time"

	"knowledge-graph/internal/domain/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository implements user.Repository using GORM.
type UserRepository struct {
	db       *gorm.DB
	roleRepo user.RoleRepository
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *gorm.DB, roleRepo user.RoleRepository) *UserRepository {
	return &UserRepository{db: db, roleRepo: roleRepo}
}

func toDomainUser(m *UserModel, role string) (*user.User, error) {
	return user.NewUser(m.ID, m.Login, m.Email, m.PasswordHash, role, m.CreatedAt, time.Time{}, m.DeletedAt)
}

func fromDomainUser(u *user.User) *UserModel {
	return &UserModel{
		ID:           u.ID(),
		Login:        u.Login(),
		Email:        u.Email(),
		PasswordHash: u.PasswordHash(),
		CreatedAt:    u.CreatedAt(),
	}
}

func (r *UserRepository) loadRole(ctx context.Context, roleID *uuid.UUID) (string, error) {
	if roleID == nil || r.roleRepo == nil {
		return "", nil
	}
	roleName, err := r.roleRepo.FindByID(ctx, *roleID)
	if errors.Is(err, user.ErrRoleNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return roleName, nil
}

// Create saves a new user.
func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	model := fromDomainUser(u)
	if u.Role() != "" && r.roleRepo != nil {
		roleID, err := r.roleRepo.FindByName(ctx, u.Role())
		if errors.Is(err, user.ErrRoleNotFound) {
			return user.ErrRoleNotFound
		}
		if err != nil {
			return err
		}
		model.RoleID = &roleID
	}
	return r.db.WithContext(ctx).Create(model).Error
}

// FindByID searches a user by ID, including soft-deleted users.
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	role, err := r.loadRole(ctx, model.RoleID)
	if err != nil {
		return nil, err
	}
	return toDomainUser(&model, role)
}

// FindByLogin searches a user by login.
func (r *UserRepository) FindByLogin(ctx context.Context, login string) (*user.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&model, "login = ?", login).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	role, err := r.loadRole(ctx, model.RoleID)
	if err != nil {
		return nil, err
	}
	return toDomainUser(&model, role)
}

// FindByEmail searches a user by email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").First(&model, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	role, err := r.loadRole(ctx, model.RoleID)
	if err != nil {
		return nil, err
	}
	return toDomainUser(&model, role)
}

// Update saves changes to an existing user.
func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", u.ID()).Error
	if err != nil {
		return err
	}
	model.Login = u.Login()
	model.Email = u.Email()
	model.PasswordHash = u.PasswordHash()
	if u.Role() != "" && r.roleRepo != nil {
		roleID, err := r.roleRepo.FindByName(ctx, u.Role())
		if err == nil {
			model.RoleID = &roleID
		}
	}
	return r.db.WithContext(ctx).Save(&model).Error
}

// Delete removes a user from the database.
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id).Error
}

// SoftDelete marks the user as deleted by setting deleted_at.
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Update("deleted_at", &now).Error
}

// EmailExists checks whether the email is already taken by another user.
func (r *UserRepository) EmailExists(ctx context.Context, email string, excludeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).
		Where("email = ? AND id != ? AND deleted_at IS NULL", email, excludeID).
		Count(&count).Error
	return count > 0, err
}

// Exists checks whether a user with the given ID exists.
func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}
