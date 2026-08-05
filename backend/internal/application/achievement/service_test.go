package achievement

import (
	"context"
	"testing"
	"time"

	achievementDomain "knowledge-graph/internal/domain/achievement"
	"knowledge-graph/internal/domain/cache/cachetest"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAchievementRepository is a mock for achievement repository
type MockAchievementRepository struct {
	mock.Mock
}

func (m *MockAchievementRepository) FindAll(ctx context.Context) ([]achievementDomain.Achievement, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]achievementDomain.Achievement), args.Error(1)
}

func (m *MockAchievementRepository) FindByCode(ctx context.Context, code string) (*achievementDomain.Achievement, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*achievementDomain.Achievement), args.Error(1)
}

func (m *MockAchievementRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]achievementDomain.Achievement, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]achievementDomain.Achievement), args.Error(1)
}

func (m *MockAchievementRepository) FindUserAchievementsWithStatus(ctx context.Context, userID uuid.UUID) ([]achievementDomain.UserAchievementWithStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]achievementDomain.UserAchievementWithStatus), args.Error(1)
}

func (m *MockAchievementRepository) SaveUserAchievement(ctx context.Context, ua achievementDomain.UserAchievement) error {
	args := m.Called(ctx, ua)
	return args.Error(0)
}

func (m *MockAchievementRepository) UserHasAchievement(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID, achievementID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAchievementRepository) MarkNotificationSeen(ctx context.Context, userID uuid.UUID, achievementID uuid.UUID) error {
	args := m.Called(ctx, userID, achievementID)
	return args.Error(0)
}

// MockEngine is a mock for achievement engine
type MockEngine struct {
	mock.Mock
}

func (m *MockEngine) Evaluate(ctx context.Context, condition achievementDomain.Condition, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, condition, userID)
	return args.Bool(0), args.Error(1)
}

func TestParseTriggerKey(t *testing.T) {
	p := parseTriggerKey("note.create")
	if p == nil {
		t.Fatal("expected non-nil map for note.create")
	}
	assert.Equal(t, "note", p["entity"])
	assert.Equal(t, "create", p["action"])

	nilp := parseTriggerKey("unknown.event")
	assert.Nil(t, nilp)
}

func TestMatchesTrigger(t *testing.T) {
	cond := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create"}
	trigger := map[string]string{"entity": "note", "action": "create"}
	ok := matchesTrigger(cond, trigger)
	assert.True(t, ok)

	cond2 := achievementDomain.Condition{Type: "count", Entity: "link", Action: "create"}
	ok2 := matchesTrigger(cond2, trigger)
	assert.False(t, ok2)
}

func TestTrackLoginAndGetStreak_WithNilRedis(t *testing.T) {
	s := NewService(nil, nil, nil, nil, nil)
	ctx := context.Background()
	uid := uuid.New()

	// Should not error when redis is nil
	err := s.TrackLogin(ctx, uid)
	assert.NoError(t, err)

	streak, err := s.GetStreak(ctx, uid)
	assert.NoError(t, err)
	assert.Equal(t, 0, streak)
}

func TestMarkNotificationSeen_EmptyID(t *testing.T) {
	s := NewService(nil, nil, nil, nil, nil)
	ctx := context.Background()
	uid := uuid.New()

	// empty achievement id should be ignored and return nil
	err := s.MarkNotificationSeen(ctx, uid, "")
	assert.NoError(t, err)
}

func TestCheckTrigger_WithMatchingAchievement(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockAchievementRepository)
	mockEngine := new(MockEngine)

	// Create a test achievement
	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	// Setup mock expectations
	mockRepo.On("FindAll", ctx).Return([]achievementDomain.Achievement{*achievement}, nil)
	mockRepo.On("UserHasAchievement", ctx, userID, achievement.ID()).Return(false, nil)
	mockEngine.On("Evaluate", ctx, condition, userID).Return(true, nil)
	mockRepo.On("SaveUserAchievement", ctx, mock.AnythingOfType("achievement.UserAchievement")).Return(nil)

	service := NewService(mockEngine, mockRepo, nil, nil, nil)

	err := service.CheckTrigger(ctx, userID, "note.create")
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
}

func TestCheckTrigger_WithAlreadyEarnedAchievement(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockAchievementRepository)
	mockEngine := new(MockEngine)

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	mockRepo.On("FindAll", ctx).Return([]achievementDomain.Achievement{*achievement}, nil)
	mockRepo.On("UserHasAchievement", ctx, userID, achievement.ID()).Return(true, nil)

	service := NewService(mockEngine, mockRepo, nil, nil, nil)

	err := service.CheckTrigger(ctx, userID, "note.create")
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEngine.AssertNotCalled(t, "Evaluate")
	mockRepo.AssertNotCalled(t, "SaveUserAchievement")
}

func TestCheckTrigger_WithUnmetCondition(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockAchievementRepository)
	mockEngine := new(MockEngine)

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 10}
	achievement, _ := achievementDomain.NewAchievement("many_notes", "Many Notes", "Description", "⭐", condition, 50, false)

	mockRepo.On("FindAll", ctx).Return([]achievementDomain.Achievement{*achievement}, nil)
	mockRepo.On("UserHasAchievement", ctx, userID, achievement.ID()).Return(false, nil)
	mockEngine.On("Evaluate", ctx, condition, userID).Return(false, nil)

	service := NewService(mockEngine, mockRepo, nil, nil, nil)

	err := service.CheckTrigger(ctx, userID, "note.create")
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockEngine.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "SaveUserAchievement")
}

func TestGetUserAchievementsWithStatus(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockAchievementRepository)

	expectedStatus := []achievementDomain.UserAchievementWithStatus{
		{
			ID:               uuid.New().String(),
			Code:             "first_note",
			NameRu:           stringPtr("Первая заметка"),
			NameEn:           stringPtr("First Note"),
			Points:           10,
			NotificationSeen: false,
		},
	}

	mockRepo.On("FindUserAchievementsWithStatus", ctx, userID).Return(expectedStatus, nil)

	service := NewService(nil, mockRepo, nil, nil, nil)

	result, err := service.GetUserAchievementsWithStatus(ctx, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, expectedStatus[0].Code, result[0].Code)
	assert.Equal(t, expectedStatus[0].Points, result[0].Points)

	mockRepo.AssertExpectations(t)
}

func TestGetAllAchievements(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockAchievementRepository)

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	expectedAchievements := []achievementDomain.Achievement{*achievement}
	mockRepo.On("FindAll", ctx).Return(expectedAchievements, nil)

	service := NewService(nil, mockRepo, nil, nil, nil)

	result, err := service.GetAllAchievements(ctx)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "first_note", result[0].Code())

	mockRepo.AssertExpectations(t)
}

func TestMarkNotificationSeen(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	achievementID := uuid.New().String()

	mockRepo := new(MockAchievementRepository)

	mockRepo.On("MarkNotificationSeen", ctx, userID, mock.AnythingOfType("uuid.UUID")).Return(nil)

	service := NewService(nil, mockRepo, nil, nil, nil)

	err := service.MarkNotificationSeen(ctx, userID, achievementID)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestSendNotification_NoSettingsNoQueue(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	s := NewService(nil, nil, nil, nil, nil)
	err := s.sendNotification(ctx, userID, *achievement)

	assert.NoError(t, err)
}

func TestParseTriggerKey_AllCases(t *testing.T) {
	cases := map[string]map[string]string{
		"note.create":    {"entity": "note", "action": "create"},
		"link.create":    {"entity": "link", "action": "create"},
		"login":          {"action": "login"},
		"search.execute": {"entity": "search", "action": "execute"},
		"share.create":   {"entity": "share", "action": "create"},
	}

	for input, expected := range cases {
		assert.Equal(t, expected, parseTriggerKey(input))
	}

	assert.Nil(t, parseTriggerKey("unknown"))
}

func TestMatchesTrigger_NonCountType(t *testing.T) {
	cond := achievementDomain.Condition{Type: "streak", Action: "login"}
	trigger := map[string]string{"action": "login"}

	ok := matchesTrigger(cond, trigger)
	assert.True(t, ok)
}

func TestCheckTrigger_WithInvalidTrigger(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockAchievementRepository)
	mockEngine := new(MockEngine)

	// Mock FindAll to return empty list (no achievements match invalid trigger)
	mockRepo.On("FindAll", ctx).Return([]achievementDomain.Achievement{}, nil)

	service := NewService(mockEngine, mockRepo, nil, nil, nil)

	err := service.CheckTrigger(ctx, userID, "invalid.trigger")
	assert.Error(t, err) // Should return error for invalid triggers
	assert.Contains(t, err.Error(), "invalid trigger key")

	mockRepo.AssertExpectations(t)
	mockEngine.AssertNotCalled(t, "Evaluate")
}

func TestCheckStreaks(t *testing.T) {
	ctx := context.Background()

	mockRepo := new(MockAchievementRepository)

	condition := achievementDomain.Condition{Type: "streak", Action: "login", Threshold: 7}
	achievement, _ := achievementDomain.NewAchievement("streak_7", "7 Day Streak", "Description", "🔥", condition, 100, false)

	mockRepo.On("FindAll", ctx).Return([]achievementDomain.Achievement{*achievement}, nil)

	service := NewService(nil, mockRepo, nil, nil, nil)

	err := service.CheckStreaks(ctx)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

func TestTrackLoginAndGetStreak_WithRedis(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()

	mockRepo := new(MockAchievementRepository)
	mockRepo.On("FindAll", ctx).Return([]achievementDomain.Achievement{}, nil)

	service := NewService(nil, mockRepo, nil, cachetest.NewFakeCacheClient(), nil)

	err := service.TrackLogin(ctx, uid)
	assert.NoError(t, err)

	streak, err := service.GetStreak(ctx, uid)
	assert.NoError(t, err)
	assert.Equal(t, 1, streak)

	err = service.TrackLogin(ctx, uid)
	assert.NoError(t, err)

	streak, err = service.GetStreak(ctx, uid)
	assert.NoError(t, err)
	assert.Equal(t, 2, streak)
}

func TestGetUserAchievements(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	mockRepo := new(MockAchievementRepository)
	mockRepo.On("FindByUserID", ctx, uid).Return([]achievementDomain.Achievement{*achievement}, nil)

	service := NewService(nil, mockRepo, nil, nil, nil)
	result, err := service.GetUserAchievements(ctx, uid)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "first_note", result[0].Code())
}

type mockTaskQueue struct {
	mock.Mock
}

func (m *mockTaskQueue) EnqueueBackupToCloud(ctx context.Context, localPath, remoteKey, backupDate string) error {
	return m.Called(ctx, localPath, remoteKey, backupDate).Error(0)
}

func (m *mockTaskQueue) EnqueueRefreshRecommendations(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return m.Called(ctx, noteID, delay).Error(0)
}

func (m *mockTaskQueue) EnqueueExtractKeywords(ctx context.Context, noteID string, topN int) error {
	return m.Called(ctx, noteID, topN).Error(0)
}

func (m *mockTaskQueue) EnqueueComputeEmbedding(ctx context.Context, noteID string) error {
	return m.Called(ctx, noteID).Error(0)
}

func (m *mockTaskQueue) EnqueueRecalculateLinkWeights(ctx context.Context, noteID uuid.UUID, delay time.Duration) error {
	return m.Called(ctx, noteID, delay).Error(0)
}

func (m *mockTaskQueue) EnqueueNotification(ctx context.Context, payload []byte) error {
	return m.Called(ctx, payload).Error(0)
}

func (m *mockTaskQueue) EnqueueImportBookmarks(ctx context.Context, userID uuid.UUID, taskID string, items []byte) error {
	return nil
}

func TestSendNotification_WithTaskQueue(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	tq := new(mockTaskQueue)
	tq.On("EnqueueNotification", ctx, mock.AnythingOfType("[]uint8")).Return(nil)

	s := NewService(nil, nil, nil, nil, tq)
	err := s.sendNotification(ctx, userID, *achievement)

	assert.NoError(t, err)
	tq.AssertExpectations(t)
}

func TestSendNotification_TaskQueueError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	tq := new(mockTaskQueue)
	tq.On("EnqueueNotification", ctx, mock.AnythingOfType("[]uint8")).Return(assert.AnError)

	s := NewService(nil, nil, nil, nil, tq)
	err := s.sendNotification(ctx, userID, *achievement)

	assert.Error(t, err)
}

func TestCheckTrigger_FindAllError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockAchievementRepository)
	mockRepo.On("FindAll", ctx).Return(nil, assert.AnError)

	service := NewService(nil, mockRepo, nil, nil, nil)

	err := service.CheckTrigger(ctx, userID, "note.create")
	assert.Error(t, err)
}

func TestCheckTrigger_UserHasAchievementError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	mockRepo := new(MockAchievementRepository)
	mockEngine := new(MockEngine)

	condition := achievementDomain.Condition{Type: "count", Entity: "note", Action: "create", Threshold: 1}
	achievement, _ := achievementDomain.NewAchievement("first_note", "First Note", "Description", "⭐", condition, 10, false)

	mockRepo.On("FindAll", ctx).Return([]achievementDomain.Achievement{*achievement}, nil)
	mockRepo.On("UserHasAchievement", ctx, userID, achievement.ID()).Return(false, assert.AnError)

	service := NewService(mockEngine, mockRepo, nil, nil, nil)

	err := service.CheckTrigger(ctx, userID, "note.create")
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}
