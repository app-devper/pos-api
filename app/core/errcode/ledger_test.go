package errcode

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/data/ledger"

	"github.com/gin-gonic/gin"
)

func TestAbortLedgerTellsTheCallerWhatToDo(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"a rule the event breaks: 400 with the reason", &ledger.Rejected{Reason: "สต็อกไม่พอ"}, http.StatusBadRequest, "สต็อกไม่พอ"},
		{"a document the event names is missing: 404", fmt.Errorf("%w: receive x", ledger.ErrNotFound), http.StatusNotFound, SY_NOT_FOUND_002},
		{"Stock moved under it: 409, resending is safe", ledger.ErrConflict, http.StatusConflict, SY_CONFLICT_001},
		{"anything else: 400 with the feature's code", errors.New("boom"), http.StatusBadRequest, "RC-400-002"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			AbortLedger(ctx, c.err, "RC-400-002")

			if w.Code != c.status || !strings.Contains(w.Body.String(), c.body) {
				t.Fatalf("got %d %s, want %d containing %q", w.Code, w.Body.String(), c.status, c.body)
			}
		})
	}
}
