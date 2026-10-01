package comet

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"knowledge-graph/internal/domain/note"
)

func TestBuildICS_VeventAndValarm(t *testing.T) {
	due := time.Date(2026, 11, 3, 15, 30, 0, 0, time.UTC)
	var sec int64 = 900
	n := cometNote(t, uuid.New(), due, sec, nil)

	ics := BuildICS(n)

	assert.Contains(t, ics, "BEGIN:VCALENDAR\r\n")
	assert.Contains(t, ics, "BEGIN:VEVENT\r\n")
	assert.Contains(t, ics, "DTSTART:20261103T153000Z\r\n")
	assert.Contains(t, ics, "DTEND:20261103T153000Z\r\n")
	assert.Contains(t, ics, "UID:"+n.ID().String()+"@knowledge-graph\r\n")
	assert.Contains(t, ics, "BEGIN:VALARM\r\n")
	assert.Contains(t, ics, "TRIGGER:-PT900S\r\n")
	assert.True(t, strings.HasSuffix(ics, "END:VCALENDAR\r\n"))
}

func TestBuildICS_NoDueEmpty(t *testing.T) {
	title, _ := note.NewTitle("Comet")
	content, _ := note.NewContent("x")
	n := note.NewNote(title, content, note.MustType("comet"), note.Metadata{})
	assert.Equal(t, "", BuildICS(n))
}

func TestBuildICS_NoRemindNoValarm(t *testing.T) {
	due := time.Now().Add(time.Hour)
	title, _ := note.NewTitle("Comet")
	content, _ := note.NewContent("x")
	n := note.NewNote(title, content, note.MustType("comet"), note.Metadata{},
		note.WithCometFields(&due, nil, nil))

	ics := BuildICS(n)
	assert.Contains(t, ics, "BEGIN:VEVENT")
	assert.NotContains(t, ics, "VALARM")
}

func TestBuildICS_EscapesSpecialChars(t *testing.T) {
	title, _ := note.NewTitle("A;B,C")
	content, _ := note.NewContent("line1\nline2\\tail")
	n := note.NewNote(title, content, note.MustType("comet"), note.Metadata{},
		note.WithCometFields(ptrDue(), nil, nil))

	ics := BuildICS(n)
	assert.Contains(t, ics, "SUMMARY:A\\;B\\,C")
	assert.Contains(t, ics, `DESCRIPTION:line1\nline2\\tail`)
	assert.NotContains(t, ics, "line2\ntail") // raw newline would break the file
}

func TestBuildICS_LongLineFoldsAt75Octets(t *testing.T) {
	title, _ := note.NewTitle(strings.Repeat("x", 200))
	content, _ := note.NewContent("x")
	n := note.NewNote(title, content, note.MustType("comet"), note.Metadata{},
		note.WithCometFields(ptrDue(), nil, nil))

	ics := BuildICS(n)
	for _, line := range strings.Split(strings.TrimSuffix(ics, "\r\n"), "\r\n") {
		assert.LessOrEqual(t, len(line), 75, "ICS line exceeds 75 octets: %q", line)
	}
	// 200-char title must fold — continuation lines start with a space.
	assert.Contains(t, ics, "\r\n ")
}

func ptrDue() *time.Time {
	t := time.Now().Add(time.Hour)
	return &t
}
