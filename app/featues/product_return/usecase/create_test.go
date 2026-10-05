package usecase

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"strings"
	"testing"
)

type recorderStub struct {
	repositories.IProductReturn
	seen  request.ProductReturn
	calls int
	err   error
}

func (s *recorderStub) RecordProductReturn(req request.ProductReturn) (*entities.ProductReturn, error) {
	s.seen = req
	s.calls++
	return &entities.ProductReturn{}, s.err
}
func TestCreateProductReturnUsesTrustedIdentityAndMapsOutcome(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, fail := range []bool{false, true} {
		repo := &recorderStub{}
		if fail {
			repo.err = errors.New("transaction rejected")
		}
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"orderId":"order","items":[{"orderItemId":"line","quantity":1}],"BranchId":"untrusted","CreatedBy":"untrusted"}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Set("BranchId", "trusted-branch")
		ctx.Set("UserId", "trusted-user")
		CreateProductReturn(repo)(ctx)
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
func TestCreateProductReturnRejectsMalformedInputBeforeRecording(t *testing.T) {
	repo := &recorderStub{}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	CreateProductReturn(repo)(ctx)
	if w.Code != http.StatusBadRequest || repo.calls != 0 {
		t.Fatalf("invalid command recorded: status %d calls %d", w.Code, repo.calls)
	}
}
