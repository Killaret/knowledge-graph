//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"knowledge-graph/internal/domain/notification"
	"knowledge-graph/internal/domain/user"
	"knowledge-graph/internal/testutil"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// NotificationRepositoryIntegrationTestSuite covers dedupe, listing order and
// owner-scoped mark-read for the notifications table (COMET-1 stage C).
type NotificationRepositoryIntegrationTestSuite struct {
	suite.Suite
	db      *gorm.DB
	repo    *NotificationRepository
	cleanup func()
	ctx     context.Context
}

func (s *NotificationRepositoryIntegrationTestSuite) SetupSuite() {
	s.db, s.cleanup = testutil.SetupTestDB(s.T())
	s.ctx = context.Background()

	// Полный набор базовых таблиц — TruncateTables в SetupTest очищает все.
	models := []interface{}{
		&NoteModel{},
		&LinkModel{},
		&NoteKeywordModel{},
		&UserModel{},
		&TagModel{},
		&NoteTagModel{},
		&NotificationModel{},
	}
	s.Require().NoError(s.db.AutoMigrate(models...))
	// Частичный уникальный индекс из миграции 039 — AutoMigrate его не создаёт,
	// а ON CONFLICT (dedupe_key) без него не работает.
	s.Require().NoError(s.db.Exec(
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_notifications_dedupe ON notifications(dedupe_key) WHERE dedupe_key IS NOT NULL",
	).Error)
	s.repo = NewNotificationRepository(s.db)
}

func (s *NotificationRepositoryIntegrationTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *NotificationRepositoryIntegrationTestSuite) SetupTest() {
	s.Require().NoError(testutil.TruncateTables(s.db))
}

func (s *NotificationRepositoryIntegrationTestSuite) seedUser() uuid.UUID {
	id := uuid.New()
	u, err := user.NewUser(id, "u-"+id.String()[:8], "u-"+id.String()[:8]+"@example.com", "hash", "user", time.Now(), time.Now(), nil)
	s.Require().NoError(err)
	m := &UserModel{ID: u.ID(), Login: u.Login(), Email: u.Email(), PasswordHash: u.PasswordHash()}
	s.Require().NoError(s.db.Create(m).Error)
	return id
}

func (s *NotificationRepositoryIntegrationTestSuite) TestDedupeKeySuppressesDuplicate() {
	userID := s.seedUser()
	n1 := notification.NewNotification(userID, nil, notification.TypeCometReminder, "t", "b", "k1")
	n2 := notification.NewNotification(userID, nil, notification.TypeCometReminder, "t", "b", "k1")

	created, err := s.repo.CreateIfAbsent(s.ctx, n1)
	s.Require().NoError(err)
	s.True(created)

	created, err = s.repo.CreateIfAbsent(s.ctx, n2)
	s.Require().NoError(err)
	s.False(created, "same dedupe_key must not insert a second row")

	items, err := s.repo.ListByUser(s.ctx, userID, 10)
	s.Require().NoError(err)
	s.Len(items, 1)
}

func (s *NotificationRepositoryIntegrationTestSuite) TestListUnreadAndMarkRead() {
	userID := s.seedUser()
	other := s.seedUser()

	for _, key := range []string{"a", "b"} {
		n := notification.NewNotification(userID, nil, notification.TypeCometReminder, "t-"+key, "b", key)
		_, err := s.repo.CreateIfAbsent(s.ctx, n)
		s.Require().NoError(err)
	}
	foreign := notification.NewNotification(other, nil, notification.TypeCometReminder, "t", "b", "foreign")
	_, err := s.repo.CreateIfAbsent(s.ctx, foreign)
	s.Require().NoError(err)

	unread, err := s.repo.CountUnread(s.ctx, userID)
	s.Require().NoError(err)
	s.Equal(int64(2), unread)

	items, err := s.repo.ListByUser(s.ctx, userID, 10)
	s.Require().NoError(err)
	s.Len(items, 2)

	s.Require().NoError(s.repo.MarkRead(s.ctx, items[0].ID(), userID))
	unread, err = s.repo.CountUnread(s.ctx, userID)
	s.Require().NoError(err)
	s.Equal(int64(1), unread)

	// foreign notification is not visible/ownable
	s.Error(s.repo.MarkRead(s.ctx, foreign.ID(), userID))
}

func TestNotificationRepositoryIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NotificationRepositoryIntegrationTestSuite))
}
