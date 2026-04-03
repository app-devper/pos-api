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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productReceiveRepoStub struct {
	repositories.IProduct
	createProductCatalogFn func(param request.Product) (*entities.Product, error)
	createProductReceiveFn func(param request.Product) (*entities.Product, error)
}

func (s *productReceiveRepoStub) CreateProductCatalog(param request.Product) (*entities.Product, error) {
	return s.createProductCatalogFn(param)
}

func (s *productReceiveRepoStub) CreateProductReceive(param request.Product) (*entities.Product, error) {
	return s.createProductReceiveFn(param)
}

type receiveNoopStub struct {
	repositories.IReceive
}

func TestCreateProductReceiveUsesTransactionalRepositoryMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotParam request.Product
	productID := primitive.NewObjectID()
	branchID := primitive.NewObjectID().Hex()
	receiveID := primitive.NewObjectID().Hex()

	productRepo := &productReceiveRepoStub{
		createProductReceiveFn: func(param request.Product) (*entities.Product, error) {
			gotParam = param
			return &entities.Product{Id: productID, Name: param.Name}, nil
		},
	}

	body := `{"name":"Paracetamol","price":10,"costPrice":5,"unit":"TAB","quantity":2,"serialNumber":"PD-001","lotNumber":"LOT-1","expireDate":"2026-12-31T00:00:00Z","receiveId":"` + receiveID + `"}`
	req := httptest.NewRequest(http.MethodPost, "/products/receive", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateProductReceive(productRepo, &receiveNoopStub{})(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotParam.CreatedBy != "user-1" {
		t.Fatalf("expected CreatedBy user-1, got %s", gotParam.CreatedBy)
	}
	if gotParam.BranchId != branchID {
		t.Fatalf("expected BranchId %s, got %s", branchID, gotParam.BranchId)
	}
	if gotParam.ReceiveId != receiveID {
		t.Fatalf("expected ReceiveId %s, got %s", receiveID, gotParam.ReceiveId)
	}
}

func TestCreateProductUsesTransactionalCatalogMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotParam request.Product
	productID := primitive.NewObjectID()
	branchID := primitive.NewObjectID().Hex()

	productRepo := &productReceiveRepoStub{
		createProductCatalogFn: func(param request.Product) (*entities.Product, error) {
			gotParam = param
			return &entities.Product{Id: productID, Name: param.Name}, nil
		},
	}

	body := `{"name":"Ibuprofen","nameEn":"Ibuprofen","description":"pain relief","price":20,"costPrice":10,"unit":"TAB","serialNumber":"PD-002","category":"MED","status":"ACTIVE","minStock":5,"drugRegistrations":["A-1"]}`
	req := httptest.NewRequest(http.MethodPost, "/products", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-2")
	ctx.Set("BranchId", branchID)

	CreateProduct(productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotParam.CreatedBy != "user-2" {
		t.Fatalf("expected CreatedBy user-2, got %s", gotParam.CreatedBy)
	}
	if gotParam.BranchId != branchID {
		t.Fatalf("expected BranchId %s, got %s", branchID, gotParam.BranchId)
	}
	if gotParam.Quantity != 0 {
		t.Fatalf("expected Quantity 0, got %d", gotParam.Quantity)
	}
	if gotParam.SerialNumber != "PD-002" {
		t.Fatalf("expected SerialNumber PD-002, got %s", gotParam.SerialNumber)
	}
	if gotParam.MinStock != 5 {
		t.Fatalf("expected MinStock 5, got %d", gotParam.MinStock)
	}
}
