package usecase

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"pos/app/data/entities"
	"pos/app/domain/request"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type recorderStub struct {
	seen  request.StockAdjustment
	calls int
	err   error
}

func (s *recorderStub) Adjust(_ context.Context, req request.StockAdjustment) (*entities.StockAdjustment, error) {
	s.seen = req
	s.calls++
	return &entities.StockAdjustment{}, s.err
}
func TestCreateStockAdjustmentUsesTrustedIdentityAndMapsOutcome(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, fail := range []bool{false, true} {
		repo := &recorderStub{}
		if fail {
			repo.err = errors.New("transaction rejected")
		}
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"productId":"product","stockId":"stock","reason":"other","delta":1,"BranchId":"untrusted","CreatedBy":"untrusted"}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set("BranchId", "trusted-branch")
		ctx.Set("UserId", "trusted-user")
		CreateStockAdjustment(repo)(ctx)
		want := http.StatusOK
		if fail {
			want = http.StatusBadRequest
		}
		if w.Code != want {
			t.Fatalf("status %d, want %d: %s", w.Code, want, w.Body.String())
		}
		if repo.calls != 1 || repo.seen.BranchId != "trusted-branch" || repo.seen.CreatedBy != "trusted-user" {
			t.Fatalf("wrong command identity: %+v", repo.seen)
		}
	}
}
func TestCreateStockAdjustmentRejectsMalformedInputBeforeRecording(t *testing.T) {
	repo := &recorderStub{}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	CreateStockAdjustment(repo)(ctx)
	if w.Code != http.StatusBadRequest || repo.calls != 0 {
		t.Fatalf("invalid command recorded: status %d calls %d", w.Code, repo.calls)
	}
}
