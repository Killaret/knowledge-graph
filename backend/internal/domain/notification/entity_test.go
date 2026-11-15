package notification

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNotification_MarkReadIsIdempotent(t *testing.T) {
	n := NewNotification(uuid.New(), nil, TypeCometReminder, "t", "b", "k")
	assert.False(t, n.IsRead())
	n.MarkRead()
	require := n.ReadAt()
	assert.NotNil(t, require)
	n.MarkRead()
	assert.Equal(t, require, n.ReadAt(), "second MarkRead must not move the timestamp")
}
