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

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type deleteReceiveRepoStub struct {
	repositories.IReceive
	removeByIDFn func(id string) (*entities.Receive, error)
}

func (s *deleteReceiveRepoStub) RemoveReceiveById(id string) (*entities.Receive, error) {
	return s.removeByIDFn(id)
}

func TestDeleteReceiveByIdReturnsDeletedReceive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	var gotID string
	repo := &deleteReceiveRepoStub{
		removeByIDFn: func(id string) (*entities.Receive, error) {
			gotID = id
			return &entities.Receive{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/receives/"+receiveID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}

	DeleteReceiveById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != receiveID {
		t.Fatalf("expected receive id %s, got %s", receiveID, gotID)
	}
}

func TestDeleteReceiveByIdReturnsError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	repo := &deleteReceiveRepoStub{
		removeByIDFn: func(id string) (*entities.Receive, error) {
			return nil, errors.New("delete failed")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/receives/"+receiveID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}

	DeleteReceiveById(repo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.RC_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.RC_BAD_REQUEST_002, w.Body.String())
	}
}
