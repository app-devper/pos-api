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
	"pos/app/domain/constant"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type deleteReceiveRepoStub struct {
	repositories.IReceive
	getByIDFn          func(id string) (*entities.Receive, error)
	updateStatusByIDFn func(id string, status string, updatedBy string) (*entities.Receive, error)
}

func (s *deleteReceiveRepoStub) GetReceiveById(id string) (*entities.Receive, error) {
	return s.getByIDFn(id)
}

func (s *deleteReceiveRepoStub) UpdateReceiveStatusById(id string, status string, updatedBy string) (*entities.Receive, error) {
	return s.updateStatusByIDFn(id, status, updatedBy)
}

func TestDeleteReceiveByIdReturnsDeletedReceive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()
	var gotID string
	var gotStatus string
	var gotUpdatedBy string
	repo := &deleteReceiveRepoStub{
		getByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID}, nil
		},
		updateStatusByIDFn: func(id string, status string, updatedBy string) (*entities.Receive, error) {
			gotID = id
			gotStatus = status
			gotUpdatedBy = updatedBy
			return &entities.Receive{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/receives/"+receiveID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}
	ctx.Set("BranchId", branchID.Hex())
	ctx.Set("UserId", "user-1")

	DeleteReceiveById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != receiveID {
		t.Fatalf("expected receive id %s, got %s", receiveID, gotID)
	}
	if gotStatus != constant.CANCELLED {
		t.Fatalf("expected status %s, got %s", constant.CANCELLED, gotStatus)
	}
	if gotUpdatedBy != "user-1" {
		t.Fatalf("expected UpdatedBy user-1, got %s", gotUpdatedBy)
	}
}

func TestDeleteReceiveByIdReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()
	repo := &deleteReceiveRepoStub{
		getByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID}, nil
		},
		updateStatusByIDFn: func(id string, status string, updatedBy string) (*entities.Receive, error) {
			return nil, errors.New("delete failed")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/receives/"+receiveID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}
	ctx.Set("BranchId", branchID.Hex())
	ctx.Set("UserId", "user-1")

	DeleteReceiveById(repo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.RC_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.RC_BAD_REQUEST_002, w.Body.String())
	}
}

func TestDeleteReceiveByIdRejectsImportedReceive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()
	repo := &deleteReceiveRepoStub{
		getByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Status: constant.IMPORTED}, nil
		},
		updateStatusByIDFn: func(id string, status string, updatedBy string) (*entities.Receive, error) {
			t.Fatal("update status should not be called for imported receive")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/receives/"+receiveID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}
	ctx.Set("BranchId", branchID.Hex())
	ctx.Set("UserId", "user-1")

	DeleteReceiveById(repo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "cannot cancel imported receive") {
		t.Fatalf("expected imported receive error, got %s", w.Body.String())
	}
}
