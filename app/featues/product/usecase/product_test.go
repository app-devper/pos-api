package usecase

import (
	"errors"
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
	updateProductByIDFn    func(id string, param request.UpdateProduct) (*entities.Product, error)
	createHistoryFn        func(param request.ProductHistory) (*entities.ProductHistory, error)
	getProductAllFn        func(param request.GetProduct) ([]entities.ProductDetail, error)
	getProductByIDFn       func(id string) (*entities.Product, error)
	getUnitsByProductIDFn  func(productId string) ([]entities.ProductUnit, error)
	getPricesByProductIDFn func(productId string) ([]entities.ProductPrice, error)
	getStocksByProductIDFn func(productId string, branchId string) ([]entities.ProductStock, error)
}

func (s *productReceiveRepoStub) CreateProductCatalog(param request.Product) (*entities.Product, error) {
	return s.createProductCatalogFn(param)
}

func (s *productReceiveRepoStub) CreateProductReceive(param request.Product) (*entities.Product, error) {
	return s.createProductReceiveFn(param)
}

func (s *productReceiveRepoStub) UpdateProductById(id string, param request.UpdateProduct) (*entities.Product, error) {
	return s.updateProductByIDFn(id, param)
}

func (s *productReceiveRepoStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

func (s *productReceiveRepoStub) GetProductAll(param request.GetProduct) ([]entities.ProductDetail, error) {
	return s.getProductAllFn(param)
}

func (s *productReceiveRepoStub) GetProductById(id string) (*entities.Product, error) {
	return s.getProductByIDFn(id)
}

func (s *productReceiveRepoStub) GetProductUnitsByProductId(productId string) ([]entities.ProductUnit, error) {
	return s.getUnitsByProductIDFn(productId)
}

func (s *productReceiveRepoStub) GetProductPricesByProductId(productId string) ([]entities.ProductPrice, error) {
	return s.getPricesByProductIDFn(productId)
}

func (s *productReceiveRepoStub) GetProductStocksByProductId(productId string, branchId string) ([]entities.ProductStock, error) {
	return s.getStocksByProductIDFn(productId, branchId)
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

func TestUpdateProductByIdLogsHistoryFailureButReturnsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	historyCalled := false

	productRepo := &productReceiveRepoStub{
		updateProductByIDFn: func(id string, param request.UpdateProduct) (*entities.Product, error) {
			return &entities.Product{Id: primitive.NewObjectID(), Name: param.Name}, nil
		},
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			historyCalled = true
			return nil, errors.New("history failed")
		},
	}

	body := `{"name":"Updated Product","nameEn":"Updated Product","description":"updated","price":20,"costPrice":10,"unit":"TAB","serialNumber":"PD-002","category":"MED","status":"ACTIVE","minStock":5}`
	req := httptest.NewRequest(http.MethodPatch, "/products/"+productID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "productId", Value: productID}}
	ctx.Set("UserId", "user-2")
	ctx.Set("BranchId", branchID)

	UpdateProductById(productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !historyCalled {
		t.Fatal("expected history creation to be attempted")
	}
}

func TestGetProductsPassesBranchIdToRepository(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	var gotReq request.GetProduct

	productRepo := &productReceiveRepoStub{
		getProductAllFn: func(param request.GetProduct) ([]entities.ProductDetail, error) {
			gotReq = param
			return []entities.ProductDetail{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products?category=MED", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", branchID)

	GetProducts(productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotReq.BranchId != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotReq.BranchId)
	}
	if gotReq.Category != "MED" {
		t.Fatalf("expected category MED, got %s", gotReq.Category)
	}
}

func TestGetProductByIdUsesBranchScopedStocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := primitive.NewObjectID()
	branchID := primitive.NewObjectID().Hex()
	var gotStockProductID string
	var gotStockBranchID string

	productRepo := &productReceiveRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return &entities.Product{
				Id:           productID,
				Name:         "Paracetamol",
				SerialNumber: "PD-001",
			}, nil
		},
		getUnitsByProductIDFn: func(productId string) ([]entities.ProductUnit, error) {
			return []entities.ProductUnit{{Id: primitive.NewObjectID(), ProductId: productID, Unit: "TAB"}}, nil
		},
		getPricesByProductIDFn: func(productId string) ([]entities.ProductPrice, error) {
			return []entities.ProductPrice{{Id: primitive.NewObjectID(), ProductId: productID, Price: 10}}, nil
		},
		getStocksByProductIDFn: func(productId string, branchId string) ([]entities.ProductStock, error) {
			gotStockProductID = productId
			gotStockBranchID = branchId
			return []entities.ProductStock{{Id: primitive.NewObjectID(), ProductId: productID}}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products/"+productID.Hex(), nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "productId", Value: productID.Hex()}}
	ctx.Set("BranchId", branchID)

	GetProductById(productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotStockProductID != productID.Hex() {
		t.Fatalf("expected stock product id %s, got %s", productID.Hex(), gotStockProductID)
	}
	if gotStockBranchID != branchID {
		t.Fatalf("expected stock branch id %s, got %s", branchID, gotStockBranchID)
	}
}
