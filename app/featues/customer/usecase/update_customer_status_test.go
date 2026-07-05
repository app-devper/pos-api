package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type customerStatusRepoStub struct {
	repositories.ICustomer
	updateStatusFn func(id string, form request.UpdateCustomerStatus) (*entities.Customer, error)
}

func (s *customerStatusRepoStub) UpdateCustomerStatusById(id string, form request.UpdateCustomerStatus) (*entities.Customer, error) {
	return s.updateStatusFn(id, form)
}

func TestUpdateCustomerStatusByIdRejectsInvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &customerStatusRepoStub{
		updateStatusFn: func(id string, form request.UpdateCustomerStatus) (*entities.Customer, error) {
			t.Fatal("update should not be called when status is invalid")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/customers/"+primitive.NewObjectID().Hex()+"/status", strings.NewReader(`{"status":"DELETED"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "customerId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")

	UpdateCustomerStatusById(repo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.CU_BAD_REQUEST_001) {
		t.Fatalf("expected errcode %s, got %s", errcode.CU_BAD_REQUEST_001, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid customer status") {
		t.Fatalf("expected invalid status message, got %s", w.Body.String())
	}
}

func TestUpdateCustomerStatusByIdAcceptsArchivedStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotForm request.UpdateCustomerStatus
	repo := &customerStatusRepoStub{
		updateStatusFn: func(id string, form request.UpdateCustomerStatus) (*entities.Customer, error) {
			gotForm = form
			return &entities.Customer{Id: primitive.NewObjectID(), Status: form.Status}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/customers/"+primitive.NewObjectID().Hex()+"/status", strings.NewReader(`{"status":"ARCHIVED"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "customerId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")

	UpdateCustomerStatusById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotForm.Status != "ARCHIVED" {
		t.Fatalf("expected ARCHIVED status, got %s", gotForm.Status)
	}
	if gotForm.UpdatedBy != "user-1" {
		t.Fatalf("expected UpdatedBy user-1, got %s", gotForm.UpdatedBy)
	}
}
