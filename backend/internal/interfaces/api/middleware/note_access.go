package middleware

import (
	"net/http"

	"knowledge-graph/internal/domain/note"
	apicommon "knowledge-graph/internal/interfaces/api/common"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NoteAccessLevel selects how a /notes/:id route is authorized.
type NoteAccessLevel int

const (
	// NoteAccessRead allows the owner and, for public notes, any caller —
	// including anonymous ones once the route leaves JWT protection.
	NoteAccessRead NoteAccessLevel = iota
	// NoteAccessWrite allows only the owner, regardless of visibility.
	// Use it for mutations and for owner-scoped views (shares, drafts).
	NoteAccessWrite
)

// RequireNoteAccess enforces object-level authorization on routes carrying
// a note id in the `:id` param. Every refusal answers 404 — the existence of
// someone else's private note must not be confirmable. The SKIP_AUTH test
// bypass is honored so the E2E suite keeps working against seeded data.
func RequireNoteAccess(noteRepo note.Repository, access NoteAccessLevel) gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsSkipAuth(c.Request.Context()) {
			c.Next()
			return
		}

		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			// Not a UUID — let the handler keep producing its 400.
			c.Next()
			return
		}

		n, err := noteRepo.FindByID(c.Request.Context(), id)
		if err != nil {
			apicommon.InternalErrorWithMessage(c, apicommon.MsgFailedFetchNote)
			c.Abort()
			return
		}
		if n == nil {
			apicommon.NotFound(c, "Note")
			c.Abort()
			return
		}

		// Ownership requires an authenticated identity: the test stack's
		// seeded user legitimately authenticates as uuid.Nil, while an
		// anonymous request has no identity at all and owns nothing.
		userID, authed := GetUserID(c)
		owned := authed && n.IsOwnedBy(userID)
		allowed := owned ||
			(access == NoteAccessRead && c.Request.Method == http.MethodGet && n.IsPublic())
		if !allowed {
			apicommon.NotFound(c, "Note")
			c.Abort()
			return
		}

		c.Next()
	}
}
