package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

type supplierRepoStub struct {
	repositories.ISupplier
	createSupplierFn                   func(form request.Supplier) (*entities.Supplier, error)
	createOrUpdateSupplierByClientIDFn func(id string, form request.Supplier) (*entities.Supplier, error)
}

func (s *supplierRepoStub) CreateSupplier(form request.Supplier) (*entities.Supplier, error) {
	return s.createSupplierFn(form)
}

func (s *supplierRepoStub) CreateOrUpdateSupplierByClientId(id string, form request.Supplier) (*entities.Supplier, error) {
	return s.createOrUpdateSupplierByClientIDFn(id, form)
}

func TestCreateSupplierPassesUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotForm request.Supplier
	repo := &supplierRepoStub{
		createSupplierFn: func(form request.Supplier) (*entities.Supplier, error) {
			gotForm = form
			return &entities.Supplier{Name: form.Name}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/suppliers", strings.NewReader(`{"name":"Supplier A","address":"Bangkok"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")

	CreateSupplier(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotForm.UpdatedBy != "user-1" {
		t.Fatalf("expected UpdatedBy user-1, got %s", gotForm.UpdatedBy)
	}
}

func TestUpdateSupplierInfoPassesClientIdAndUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotClientID string
	var gotForm request.Supplier
	repo := &supplierRepoStub{
		createOrUpdateSupplierByClientIDFn: func(id string, form request.Supplier) (*entities.Supplier, error) {
			gotClientID = id
			gotForm = form
			return &entities.Supplier{ClientId: id, Name: form.Name}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/supplier-info", strings.NewReader(`{"name":"Supplier A","address":"Bangkok"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("ClientId", "client-123")

	UpdateSupplierInfo(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotClientID != "client-123" {
		t.Fatalf("expected client id client-123, got %s", gotClientID)
	}
	if gotForm.UpdatedBy != "user-1" {
		t.Fatalf("expected UpdatedBy user-1, got %s", gotForm.UpdatedBy)
	}
}
