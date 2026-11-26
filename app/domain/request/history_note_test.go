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
