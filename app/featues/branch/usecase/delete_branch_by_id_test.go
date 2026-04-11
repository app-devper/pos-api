package usecase

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type branchDeleteRepoStub struct {
	repositories.IBranch
	removeBranchFn func(id string, updatedBy string) (*entities.Branch, error)
}

func (s *branchDeleteRepoStub) RemoveBranchById(id string, updatedBy string) (*entities.Branch, error) {
	return s.removeBranchFn(id, updatedBy)
}

func TestDeleteBranchByIdPassesUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	var gotID string
	var gotUpdatedBy string

	repo := &branchDeleteRepoStub{
		removeBranchFn: func(id string, updatedBy string) (*entities.Branch, error) {
			gotID = id
			gotUpdatedBy = updatedBy
			return &entities.Branch{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/branches/"+branchID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "branchId", Value: branchID}}
	ctx.Set("UserId", "user-1")

	DeleteBranchById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotID)
	}
	if gotUpdatedBy != "user-1" {
		t.Fatalf("expected updatedBy user-1, got %s", gotUpdatedBy)
	}
}
