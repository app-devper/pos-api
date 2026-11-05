package errcode

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/data/ledger"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
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
		{"the time budget ran out: 504 with the system code", context.DeadlineExceeded, http.StatusGatewayTimeout, SY_TIMEOUT_001},
		{"the client left: 499 and no body", context.Canceled, 499, ""},
		{"anything else: 400 with the feature's code", errors.New("boom"), http.StatusBadRequest, "RC-400-002"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			AbortLedger(ctx, c.err, "RC-400-002")

			assertAnswer(t, w, c.status, c.body)
		})
	}
}

func TestAbortLookupTellsTheCallerWhatToDo(t *testing.T) {
	cases := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"no such document in this branch: 404", mongo.ErrNoDocuments, http.StatusNotFound, SY_NOT_FOUND_002},
		{"anything else: 400 with the feature's code", errors.New("boom"), http.StatusBadRequest, "RC-400-002"},
		{"the time budget ran out: 504 with the system code", context.DeadlineExceeded, http.StatusGatewayTimeout, SY_TIMEOUT_001},
		{"the client left: 499 and no body", context.Canceled, 499, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			AbortLookup(ctx, c.err, "RC-400-002")

			assertAnswer(t, w, c.status, c.body)
		})
	}
}

// assertAnswer checks the status and that the body holds want; an empty want
// means the body must be empty.
func assertAnswer(t *testing.T, w *httptest.ResponseRecorder, status int, want string) {
	t.Helper()
	body := w.Body.String()
	bodyOK := strings.Contains(body, want)
	if want == "" {
		bodyOK = body == ""
	}
	if w.Code != status || !bodyOK {
		t.Fatalf("got %d %q, want %d containing %q", w.Code, body, status, want)
	}
}
