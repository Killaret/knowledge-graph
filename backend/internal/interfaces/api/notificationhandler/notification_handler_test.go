package notificationhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"knowledge-graph/internal/domain/notification"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type fakeRepo struct {
	items     []*notification.Notification
	unread    int64
	markErr   error
	markedID  uuid.UUID
	markedUid uuid.UUID
	marked    bool
}

func (f *fakeRepo) CreateIfAbsent(ctx context.Context, n *notification.Notification) (bool, error) {
	return true, nil
}

func (f *fakeRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*notification.Notification, error) {
	return f.items, nil
}

func (f *fakeRepo) CountUnread(ctx context.Context, userID uuid.UUID) (int64, error) {
	return f.unread, nil
}

func (f *fakeRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	f.marked, f.markedID, f.markedUid = true, id, userID
	return f.markErr
}

func newCtx(t *testing.T, method, url string, userID *uuid.UUID) (*httptest.ResponseRecorder, *gin.Context) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, url, nil)
	c.Request = req
	if userID != nil {
		c.Set(middleware.ContextUserIDKey, *userID)
	}
	return w, c
}

func TestList_ReturnsItemsAndUnreadCount(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	repo := &fakeRepo{
		items: []*notification.Notification{
			notification.NewNotification(userID, &noteID, notification.TypeCometReminder, "Comet: X", "body", "key1"),
		},
		unread: 2,
	}
	h := New(repo)

	w, c := newCtx(t, http.MethodGet, "/api/v1/notifications", &userID)
	h.List(c)

	assert.Equal(t, http.StatusOK, c.Writer.Status())
	assert.Contains(t, w.Body.String(), `"unread_count":2`)
	assert.Contains(t, w.Body.String(), "Comet: X")
	assert.Contains(t, w.Body.String(), noteID.String())
}

func TestList_Unauthorized(t *testing.T) {
	h := New(&fakeRepo{})
	_, c := newCtx(t, http.MethodGet, "/api/v1/notifications", nil)
	h.List(c)
	assert.Equal(t, http.StatusUnauthorized, c.Writer.Status())
}

func TestMarkRead_MarksOwnNotification(t *testing.T) {
	userID := uuid.New()
	id := uuid.New()
	repo := &fakeRepo{}
	h := New(repo)

	_, c := newCtx(t, http.MethodPost, "/api/v1/notifications/"+id.String()+"/read", &userID)
	c.Params = gin.Params{{Key: "id", Value: id.String()}}
	h.MarkRead(c)

	assert.Equal(t, http.StatusNoContent, c.Writer.Status())
	require.True(t, repo.marked)
	assert.Equal(t, id, repo.markedID)
	assert.Equal(t, userID, repo.markedUid)
}

func TestMarkRead_NotFoundForForeignOrMissing(t *testing.T) {
	userID := uuid.New()
	repo := &fakeRepo{markErr: gorm.ErrRecordNotFound}
	h := New(repo)

	_, c := newCtx(t, http.MethodPost, "/x", &userID)
	c.Params = gin.Params{{Key: "id", Value: uuid.New().String()}}
	h.MarkRead(c)

	assert.Equal(t, http.StatusNotFound, c.Writer.Status())
}

func TestMarkRead_BadID(t *testing.T) {
	userID := uuid.New()
	h := New(&fakeRepo{})

	_, c := newCtx(t, http.MethodPost, "/x", &userID)
	c.Params = gin.Params{{Key: "id", Value: "nope"}}
	h.MarkRead(c)

	assert.Equal(t, http.StatusBadRequest, c.Writer.Status())
}
