package request

import (
	"strings"
	"unicode"
)

// historyNoteLimit is how many characters of a note a history row keeps.
const historyNoteLimit = 200

// historyNoteEllipsis marks a note cut to historyNoteLimit.
const historyNoteEllipsis = "…"

// zeroWidthJoiner glues emoji into one sequence.
const zeroWidthJoiner = '‍'

// HistoryNote is a note as a history row shows it: on one line, its runs of
// whitespace collapsed to one space, and, when longer than historyNoteLimit
// characters, cut at a character boundary at or before that limit and
// followed by historyNoteEllipsis.
func HistoryNote(note string) string {
	line := strings.Join(strings.Fields(note), " ")
	runes := []rune(line)
	if len(runes) <= historyNoteLimit {
		return line
	}
	return string(runes[:boundaryAtOrBefore(runes, historyNoteLimit)]) + historyNoteEllipsis
}

// boundaryAtOrBefore steps cut back while it would leave a mark or a variation
// selector without its base, or split a zero-width-joined sequence.
func boundaryAtOrBefore(runes []rune, cut int) int {
	for cut > 0 && (continuesCharacter(runes[cut]) || runes[cut-1] == zeroWidthJoiner) {
		cut--
	}
	return cut
}

// continuesCharacter reports whether r belongs to the character before it.
func continuesCharacter(r rune) bool {
	return unicode.In(r, unicode.Mn, unicode.Mc, unicode.Me, unicode.Variation_Selector) || r == zeroWidthJoiner
}
