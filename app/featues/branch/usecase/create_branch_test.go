package usecase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

type branchRepoStub struct {
	repositories.IBranch
	createBranchFn func(form request.Branch) (*entities.Branch, error)
}

func (s *branchRepoStub) CreateBranch(form request.Branch) (*entities.Branch, error) {
	return s.createBranchFn(form)
}

type branchSequenceRepoStub struct {
	repositories.ISequence
	nextSequenceFn func(name string) (*entities.Sequence, error)
}

func (s *branchSequenceRepoStub) NextSequence(name string) (*entities.Sequence, error) {
	return s.nextSequenceFn(name)
}

func TestCreateBranchFailsWhenSequenceLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchRepo := &branchRepoStub{
		createBranchFn: func(form request.Branch) (*entities.Branch, error) {
			t.Fatal("create branch should not be called when sequence fails")
			return nil, nil
		},
	}
	sequenceRepo := &branchSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return nil, errors.New("sequence failed")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/branches", strings.NewReader(`{"name":"HQ"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")

	CreateBranch(branchRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.BR_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s, got %s", errcode.BR_BAD_REQUEST_002, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "sequence failed") {
		t.Fatalf("expected sequence failure in response, got %s", w.Body.String())
	}
}

func TestCreateBranchFailsWhenSequenceIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchRepo := &branchRepoStub{
		createBranchFn: func(form request.Branch) (*entities.Branch, error) {
			t.Fatal("create branch should not be called when sequence is missing")
			return nil, nil
		},
	}
	sequenceRepo := &branchSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) { return nil, nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/branches", strings.NewReader(`{"name":"HQ"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")

	CreateBranch(branchRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "branch sequence not available") {
		t.Fatalf("expected missing sequence error, got %s", w.Body.String())
	}
}
