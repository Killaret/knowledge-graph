package comet

import (
	"fmt"
	"strings"

	"knowledge-graph/internal/domain/note"
)

// BuildICS renders a note as a single-event iCalendar (RFC 5545) document.
// The comet due date becomes the event start; remind_before_seconds becomes a
// VALARM display trigger. Returns an empty string when the note has no due date.
func BuildICS(n *note.Note) string {
	due := n.DueAt()
	if due == nil {
		return ""
	}
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:-//knowledge-graph//comet//EN",
		"BEGIN:VEVENT",
		"UID:" + n.ID().String() + "@knowledge-graph",
		"DTSTAMP:" + n.CreatedAt().UTC().Format("20060102T150405Z"),
		"DTSTART:" + due.UTC().Format("20060102T150405Z"),
		"DTEND:" + due.UTC().Format("20060102T150405Z"),
		"SUMMARY:" + escapeICS(n.Title().String()),
	}
	if body := strings.TrimSpace(n.Content().String()); body != "" {
		lines = append(lines, "DESCRIPTION:"+escapeICS(body))
	}
	if off := n.RemindBeforeSeconds(); off != nil && *off > 0 {
		lines = append(lines,
			"BEGIN:VALARM",
			"ACTION:DISPLAY",
			fmt.Sprintf("TRIGGER:-PT%dS", *off),
			"DESCRIPTION:"+escapeICS(n.Title().String()),
			"END:VALARM",
		)
	}
	lines = append(lines, "END:VEVENT", "END:VCALENDAR")
	var b strings.Builder
	for _, l := range lines {
		foldICSLine(&b, l)
	}
	return b.String()
}

// escapeICS applies RFC 5545 §3.3.11 text escaping.
func escapeICS(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, ";", `\;`)
	s = strings.ReplaceAll(s, ",", `\,`)
	s = strings.ReplaceAll(s, "\r\n", `\n`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	return s
}

// foldICSLine writes one logical line split into ≤75-octet physical lines
// (continuation lines start with a space), terminated by CRLF.
func foldICSLine(b *strings.Builder, line string) {
	first := true
	for len(line) > 0 {
		if !first {
			b.WriteString(" ")
		}
		limit := 75
		if !first {
			limit = 74
		}
		if len(line) <= limit {
			b.WriteString(line)
			break
		}
		cut := limit
		for cut > 0 && (line[cut]&0xC0) == 0x80 {
			cut--
		}
		if cut == 0 {
			cut = limit
		}
		b.WriteString(line[:cut])
		b.WriteString("\r\n")
		line = line[cut:]
		first = false
	}
	b.WriteString("\r\n")
}
