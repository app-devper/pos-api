package app

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// A usecase that answers a repository error with a bare 400 hides what went
// wrong: a missing document, a timed-out database and a duplicate key all read
// as the caller's fault. Those answers belong to errcode.AbortLookup,
// errcode.AbortWrite and errcode.AbortLedger, which know the difference.
var directRepositoryAbort = regexp.MustCompile(`errcode\.Abort\(\s*ctx,\s*http\.StatusBadRequest,\s*errcode\.\w+_BAD_REQUEST_002,\s*err\.Error\(\)\s*\)`)

func TestRepositoryErrorsAnswerThroughTheSharedHelpers(t *testing.T) {
	var offenders []string
	err := filepath.WalkDir("featues", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(src), "\n") {
			if directRepositoryAbort.MatchString(line) {
				offenders = append(offenders, path+": "+strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) > 0 {
		t.Fatalf("%d usecase(s) answer a repository error with a bare 400; use errcode.AbortLookup (reads), AbortWrite (writes) or AbortLedger:\n%s", len(offenders), strings.Join(offenders, "\n"))
	}
}
