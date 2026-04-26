//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// UserRepositoryIntegrationTestSuite - интеграционные тесты для UserRepository
type UserRepositoryIntegrationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	repo    *UserRepository
	cleanup func()
	ctx     context.Context
}

func (s *UserRepositoryIntegrationTestSuite) SetupSuite() {
	s.db, s.cleanup = testutil.SetupTestDB(s.T())
	s.ctx = context.Background()

	// Миграция всех моделей (для корректной работы TruncateTables)
	models := []interface{}{
		&NoteModel{},
		&LinkModel{},
		&NoteKeywordModel{},
		&UserModel{},
		&TagModel{},
		&NoteTagModel{},
	}
	err := s.db.AutoMigrate(models...)
	s.Require().NoError(err, "failed to migrate models")

	// Создаем репозиторий
	s.repo = NewUserRepository(s.db)
}

func (s *UserRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *UserRepositoryIntegrationTestSuite) SetupTest() {
	// Очищаем таблицы перед каждым тестом
	err := testutil.TruncateTables(s.db)
	s.Require().NoError(err, "failed to truncate tables")
}

// TestCreate - создание пользователя
func (s *UserRepositoryIntegrationTestSuite) TestCreate() {
	user := &UserModel{
		ID:           uuid.New(),
		Login:        "testuser",
		PasswordHash: "hashed_password_123",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Проверяем что пользователь создан
	found, err := s.repo.FindByID(s.ctx, user.ID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal("testuser", found.Login)
	s.Equal("hashed_password_123", found.PasswordHash)
	s.Equal("user", found.Role)
}

// TestFindByID - поиск по ID
func (s *UserRepositoryIntegrationTestSuite) TestFindByID() {
	user := &UserModel{
		ID:           uuid.New(),
		Login:        "findbyid_user",
		PasswordHash: "hash123",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Находим по ID
	found, err := s.repo.FindByID(s.ctx, user.ID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID, found.ID)
	s.Equal("findbyid_user", found.Login)
}

// TestFindByID_NotFound - поиск несуществующего
func (s *UserRepositoryIntegrationTestSuite) TestFindByID_NotFound() {
	found, err := s.repo.FindByID(s.ctx, uuid.New())
	s.NoError(err)
	s.Nil(found)
}

// TestFindByLogin - поиск по логину
func (s *UserRepositoryIntegrationTestSuite) TestFindByLogin() {
	user := &UserModel{
		ID:           uuid.New(),
		Login:        "john_doe",
		PasswordHash: "secure_hash",
		Role:         "admin",
		CreatedAt:    time.Now(),
	}

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Находим по логину
	found, err := s.repo.FindByLogin(s.ctx, "john_doe")
	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID, found.ID)
	s.Equal("john_doe", found.Login)
	s.Equal("admin", found.Role)
}

// TestFindByLogin_NotFound - поиск несуществующего логина
func (s *UserRepositoryIntegrationTestSuite) TestFindByLogin_NotFound() {
	found, err := s.repo.FindByLogin(s.ctx, "nonexistent_user")
	s.NoError(err)
	s.Nil(found)
}

// TestDuplicateLogin - проверка уникальности логина
func (s *UserRepositoryIntegrationTestSuite) TestDuplicateLogin() {
	user1 := &UserModel{
		ID:           uuid.New(),
		Login:        "unique_user",
		PasswordHash: "hash1",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	err := s.repo.Create(s.ctx, user1)
	s.NoError(err)

	// Пытаемся создать пользователя с тем же логином
	user2 := &UserModel{
		ID:           uuid.New(),
		Login:        "unique_user",
		PasswordHash: "hash2",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	err = s.repo.Create(s.ctx, user2)
	s.Error(err)
	s.Contains(strings.ToLower(err.Error()), "duplicate")
}

// TestUpdate - обновление пользователя
func (s *UserRepositoryIntegrationTestSuite) TestUpdate() {
	user := &UserModel{
		ID:           uuid.New(),
		Login:        "update_me",
		PasswordHash: "old_hash",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Обновляем через прямое изменение в БД
	err = s.db.Model(&UserModel{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"password_hash": "new_hash",
		"role":          "admin",
	}).Error
	s.NoError(err)

	// Проверяем обновление
	found, err := s.repo.FindByID(s.ctx, user.ID)
	s.NoError(err)
	s.Equal("new_hash", found.PasswordHash)
	s.Equal("admin", found.Role)
	// Логин не должен измениться
	s.Equal("update_me", found.Login)
}

// TestDelete - удаление пользователя
func (s *UserRepositoryIntegrationTestSuite) TestDelete() {
	user := &UserModel{
		ID:           uuid.New(),
		Login:        "delete_me",
		PasswordHash: "hash",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Проверяем что существует
	exists, err := s.repo.Exists(s.ctx, user.ID)
	s.NoError(err)
	s.True(exists)

	// Удаляем
	err = s.repo.Delete(s.ctx, user.ID)
	s.NoError(err)

	// Проверяем что не существует
	exists, err = s.repo.Exists(s.ctx, user.ID)
	s.NoError(err)
	s.False(exists)

	// Проверяем что FindByID возвращает nil
	found, err := s.repo.FindByID(s.ctx, user.ID)
	s.NoError(err)
	s.Nil(found)
}

// TestExists - проверка существования
func (s *UserRepositoryIntegrationTestSuite) TestExists() {
	user := &UserModel{
		ID:           uuid.New(),
		Login:        "exists_test",
		PasswordHash: "hash",
		Role:         "user",
		CreatedAt:    time.Now(),
	}

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Существует
	exists, err := s.repo.Exists(s.ctx, user.ID)
	s.NoError(err)
	s.True(exists)

	// Не существует
	exists, err = s.repo.Exists(s.ctx, uuid.New())
	s.NoError(err)
	s.False(exists)
}

// TestMultipleUsers - работа с несколькими пользователями
func (s *UserRepositoryIntegrationTestSuite) TestMultipleUsers() {
	// Создаем несколько пользователей
	for i := 0; i < 3; i++ {
		user := &UserModel{
			ID:           uuid.New(),
			Login:        "multi_user_" + string(rune('a'+i)),
			PasswordHash: "hash",
			Role:         "user",
			CreatedAt:    time.Now(),
		}
		err := s.repo.Create(s.ctx, user)
		s.NoError(err)
	}

	// Проверяем что все созданы
	var count int64
	err := s.db.Model(&UserModel{}).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(3), count)
}

// Запускаем тесты
func TestUserRepositoryIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
	suite.Run(t, new(UserRepositoryIntegrationTestSuite))
}
