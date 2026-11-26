package request

import (
	"strings"
	"testing"
)

func TestHistoryNoteIsOneShortLine(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"", ""},
		{" \n ", ""},
		{"นับใหม่", "นับใหม่"},
		{"a\nb\r\n\tc  d", "a b c d"},
		{strings.Repeat("ก", historyNoteLimit), strings.Repeat("ก", historyNoteLimit)},
		{strings.Repeat("ก", historyNoteLimit+1), strings.Repeat("ก", historyNoteLimit) + "…"},
	} {
		if got := HistoryNote(c.in); got != c.want {
			t.Fatalf("HistoryNote(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// A cut never leaves a mark without its letter or splits an emoji sequence.
func TestHistoryNoteCutsAtACharacterBoundary(t *testing.T) {
	ko := strings.Repeat("ก", historyNoteLimit)
	family := "\U0001F468\u200D\U0001F469\u200D\U0001F467"
	for _, c := range []struct{ name, in, want string }{
		{"tone mark after the cut", ko + "\u0E48" + strings.Repeat("ก", 4), strings.Repeat("ก", historyNoteLimit-1) + historyNoteEllipsis},
		{"vowel and tone after the cut", strings.Repeat("ก", historyNoteLimit-1) + "\u0E35\u0E48" + "ขขขข", strings.Repeat("ก", historyNoteLimit-2) + historyNoteEllipsis},
		{"inside a ZWJ sequence", strings.Repeat("a", historyNoteLimit-2) + family + "bbb", strings.Repeat("a", historyNoteLimit-2) + historyNoteEllipsis},
		{"variation selector after the cut", strings.Repeat("a", historyNoteLimit-1) + "\u2764\uFE0F" + "bbb", strings.Repeat("a", historyNoteLimit-1) + historyNoteEllipsis},
		{"clean cut", ko + "ขขขขข", ko + historyNoteEllipsis},
	} {
		got := HistoryNote(c.in)
		if got != c.want {
			t.Fatalf("%s: got %q (%d runes), want %q", c.name, got, len([]rune(got)), c.want)
		}
		if n := len([]rune(strings.TrimSuffix(got, historyNoteEllipsis))); n > historyNoteLimit {
			t.Fatalf("%s: kept %d runes, over %d", c.name, n, historyNoteLimit)
		}
	}
}
