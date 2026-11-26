package request

import "strings"

// historyNoteLimit is how many characters of a note a history row keeps.
const historyNoteLimit = 200

// historyNoteEllipsis marks a note cut to historyNoteLimit.
const historyNoteEllipsis = "…"

// HistoryNote is a note as a history row shows it: on one line, its runs of
// whitespace collapsed to one space, and cut to historyNoteLimit characters
// followed by historyNoteEllipsis when longer.
func HistoryNote(note string) string {
	line := strings.Join(strings.Fields(note), " ")
	runes := []rune(line)
	if len(runes) <= historyNoteLimit {
		return line
	}
	return string(runes[:historyNoteLimit]) + historyNoteEllipsis
}
