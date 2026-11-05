package errcode

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func TestRejectCarriesTheRuleMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	Reject(ctx, "PM-400-002", errors.New("ยอดซื้อไม่ถึงขั้นต่ำ"))

	assertAnswer(t, w, http.StatusBadRequest, "ยอดซื้อไม่ถึงขั้นต่ำ")
}

func TestAbortWriteTellsTheCallerWhatToDo(t *testing.T) {
	duplicate := mongo.WriteException{WriteErrors: mongo.WriteErrors{{Code: 11000, Message: "E11000 duplicate key"}}}
	cases := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{"the time budget ran out: 504 with the system code", context.DeadlineExceeded, http.StatusGatewayTimeout, SY_TIMEOUT_001},
		{"the client left: 499 and no body", context.Canceled, statusClientClosedRequest, ""},
		{"the document already exists: 409, resending is not safe", duplicate, http.StatusConflict, SY_CONFLICT_002},
		{"the document to change does not exist: 404", mongo.ErrNoDocuments, http.StatusNotFound, SY_NOT_FOUND_002},
		{"the Receive is no longer a draft: 409 with the reason", fmt.Errorf("update: %w", repositories.ErrReceiveLocked), http.StatusConflict, SY_CONFLICT_003},
		{"a counter lost its creation race every time: 503, resending is safe", fmt.Errorf("sequence RC: %w", repositories.ErrSequenceContended), http.StatusServiceUnavailable, SY_BUSY_001},
		{"anything else: 400 with the feature's code", errors.New("boom"), http.StatusBadRequest, "PM-400-002"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)

			AbortWrite(ctx, c.err, "PM-400-002")

			assertAnswer(t, w, c.status, c.body)
		})
	}
}
