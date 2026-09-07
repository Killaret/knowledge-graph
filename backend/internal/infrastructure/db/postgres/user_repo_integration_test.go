//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	domainuser "knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func newTestUser(id uuid.UUID, login, passwordHash, role string, createdAt time.Time) *domainuser.User {
	u, err := domainuser.NewUser(id, login, login+"@example.com", passwordHash, role, createdAt, time.Time{}, nil)
	if err != nil {
		panic(err)
	}
	return u
}

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
		&UserRoleModel{},
		&TagModel{},
		&NoteTagModel{},
	}
	err := s.db.AutoMigrate(models...)
	s.Require().NoError(err, "failed to migrate models")

	// Создаем репозиторий
	s.repo = NewUserRepository(s.db, nil)
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
	user := newTestUser(uuid.New(), "testuser", "hashed_password_123", "", time.Now())

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Проверяем что пользователь создан
	found, err := s.repo.FindByID(s.ctx, user.ID())
	s.NoError(err)
	s.NotNil(found)
	s.Equal("testuser", found.Login())
	s.Equal("hashed_password_123", found.PasswordHash())
	s.Empty(found.Role())
}

// TestFindByID - поиск по ID
func (s *UserRepositoryIntegrationTestSuite) TestFindByID() {
	user := newTestUser(uuid.New(), "findbyid_user", "hash123", "", time.Now())

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Находим по ID
	found, err := s.repo.FindByID(s.ctx, user.ID())
	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID(), found.ID())
	s.Equal("findbyid_user", found.Login())
}

// TestFindByID_NotFound - поиск несуществующего
func (s *UserRepositoryIntegrationTestSuite) TestFindByID_NotFound() {
	found, err := s.repo.FindByID(s.ctx, uuid.New())
	s.NoError(err)
	s.Nil(found)
}

// TestFindByLogin - поиск по логину
func (s *UserRepositoryIntegrationTestSuite) TestFindByLogin() {
	user := newTestUser(uuid.New(), "john_doe", "secure_hash", "", time.Now())

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Находим по логину
	found, err := s.repo.FindByLogin(s.ctx, "john_doe")
	s.NoError(err)
	s.NotNil(found)
	s.Equal(user.ID(), found.ID())
	s.Equal("john_doe", found.Login())
	s.Empty(found.Role())
}

// TestFindByLogin_NotFound - поиск несуществующего логина
func (s *UserRepositoryIntegrationTestSuite) TestFindByLogin_NotFound() {
	found, err := s.repo.FindByLogin(s.ctx, "nonexistent_user")
	s.NoError(err)
	s.Nil(found)
}

// TestDuplicateLogin - проверка уникальности логина
func (s *UserRepositoryIntegrationTestSuite) TestDuplicateLogin() {
	user1 := newTestUser(uuid.New(), "unique_user", "hash1", "", time.Now())

	err := s.repo.Create(s.ctx, user1)
	s.NoError(err)

	// Пытаемся создать пользователя с тем же логином
	user2 := newTestUser(uuid.New(), "unique_user", "hash2", "", time.Now())

	err = s.repo.Create(s.ctx, user2)
	s.Error(err)
	s.Contains(strings.ToLower(err.Error()), "duplicate")
}

// TestUpdate - обновление пользователя
func (s *UserRepositoryIntegrationTestSuite) TestUpdate() {
	user := newTestUser(uuid.New(), "update_me", "old_hash", "", time.Now())

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Обновляем через прямое изменение в БД
	err = s.db.Model(&UserModel{}).Where("id = ?", user.ID()).Updates(map[string]interface{}{
		"password_hash": "new_hash",
	}).Error
	s.NoError(err)

	// Проверяем обновление
	found, err := s.repo.FindByID(s.ctx, user.ID())
	s.NoError(err)
	s.Equal("new_hash", found.PasswordHash())
	s.Empty(found.Role())
	// Логин не должен измениться
	s.Equal("update_me", found.Login())
}

// TestDelete - удаление пользователя
func (s *UserRepositoryIntegrationTestSuite) TestDelete() {
	user := newTestUser(uuid.New(), "delete_me", "hash", "", time.Now())

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Проверяем что существует
	exists, err := s.repo.Exists(s.ctx, user.ID())
	s.NoError(err)
	s.True(exists)

	// Удаляем
	err = s.repo.Delete(s.ctx, user.ID())
	s.NoError(err)

	// Проверяем что не существует
	exists, err = s.repo.Exists(s.ctx, user.ID())
	s.NoError(err)
	s.False(exists)

	// Проверяем что FindByID возвращает nil
	found, err := s.repo.FindByID(s.ctx, user.ID())
	s.NoError(err)
	s.Nil(found)
}

// TestExists - проверка существования
func (s *UserRepositoryIntegrationTestSuite) TestExists() {
	user := newTestUser(uuid.New(), "exists_test", "hash", "", time.Now())

	err := s.repo.Create(s.ctx, user)
	s.NoError(err)

	// Существует
	exists, err := s.repo.Exists(s.ctx, user.ID())
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
		user := newTestUser(uuid.New(), "multi_user_"+string(rune('a'+i)), "hash", "", time.Now())
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
