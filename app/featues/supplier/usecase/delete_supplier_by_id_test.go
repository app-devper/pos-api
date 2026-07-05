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

type supplierDeleteRepoStub struct {
	repositories.ISupplier
	removeSupplierFn func(id string, updatedBy string) (*entities.Supplier, error)
}

func (s *supplierDeleteRepoStub) RemoveSupplierById(id string, updatedBy string) (*entities.Supplier, error) {
	return s.removeSupplierFn(id, updatedBy)
}

func TestDeleteSupplierByIdPassesUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	supplierID := primitive.NewObjectID().Hex()
	var gotID string
	var gotUpdatedBy string

	repo := &supplierDeleteRepoStub{
		removeSupplierFn: func(id string, updatedBy string) (*entities.Supplier, error) {
			gotID = id
			gotUpdatedBy = updatedBy
			return &entities.Supplier{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/suppliers/"+supplierID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "supplierId", Value: supplierID}}
	ctx.Set("UserId", "user-1")

	DeleteSupplierById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != supplierID {
		t.Fatalf("expected supplier id %s, got %s", supplierID, gotID)
	}
	if gotUpdatedBy != "user-1" {
		t.Fatalf("expected updatedBy user-1, got %s", gotUpdatedBy)
	}
}
