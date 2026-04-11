package usecase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productStockRepoStub struct {
	repositories.IProductStock
	createStockFn    func(param request.ProductStock) (*entities.ProductStock, error)
	getStockByIDFn   func(id string) (*entities.ProductStock, error)
	getBalanceFn     func(productId string, unitId string, branchId string) int
	createHistoryFn  func(param request.ProductHistory) (*entities.ProductHistory, error)
	updateStockFn    func(id string, param request.UpdateProductStock) (*entities.ProductStock, error)
	updateQtyFn      func(id string, quantity int) (*entities.ProductStock, error)
	removeStockFn    func(id string) (*entities.ProductStock, error)
	updateSequenceFn func(param request.UpdateProductStockSequence) ([]entities.ProductStock, error)
}

type productStockProductStub struct {
	repositories.IProduct
	getUnitByIDFn func(id string) (*entities.ProductUnit, error)
}

func (s *productStockRepoStub) CreateProductStock(param request.ProductStock) (*entities.ProductStock, error) {
	return s.createStockFn(param)
}

func (s *productStockRepoStub) GetProductStockById(id string) (*entities.ProductStock, error) {
	return s.getStockByIDFn(id)
}

func (s *productStockProductStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return s.getUnitByIDFn(id)
}

func (s *productStockRepoStub) GetProductStockBalance(productId string, unitId string, branchId string) int {
	return s.getBalanceFn(productId, unitId, branchId)
}

func (s *productStockRepoStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

func (s *productStockRepoStub) UpdateProductStockById(id string, param request.UpdateProductStock) (*entities.ProductStock, error) {
	return s.updateStockFn(id, param)
}

func (s *productStockRepoStub) UpdateProductStockQuantityById(id string, quantity int) (*entities.ProductStock, error) {
	return s.updateQtyFn(id, quantity)
}

func (s *productStockRepoStub) RemoveProductStockById(id string) (*entities.ProductStock, error) {
	return s.removeStockFn(id)
}

func (s *productStockRepoStub) UpdateProductStockSequence(param request.UpdateProductStockSequence) ([]entities.ProductStock, error) {
	return s.updateSequenceFn(param)
}

func TestCreateProductStockUsesBranchScopedBalance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	branchID := primitive.NewObjectID().Hex()
	var gotBalanceBranchID string
	var gotHistoryBranchID string

	stockRepo := &productStockRepoStub{
		createStockFn: func(param request.ProductStock) (*entities.ProductStock, error) {
			return &entities.ProductStock{
				Id:        primitive.NewObjectID(),
				ProductId: productID,
				UnitId:    unitID,
			}, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int {
			gotBalanceBranchID = branchId
			return 9
		},
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			gotHistoryBranchID = param.BranchId
			return &entities.ProductHistory{}, nil
		},
	}
	productRepo := &productStockProductStub{
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
		},
	}

	body := `{"productId":"` + productID.Hex() + `","unitId":"` + unitID.Hex() + `","quantity":2,"expireDate":"` + time.Now().UTC().Format(time.RFC3339) + `","importDate":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/product-stocks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateProductStock(stockRepo, productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotBalanceBranchID != branchID {
		t.Fatalf("expected balance lookup to use branch %s, got %s", branchID, gotBalanceBranchID)
	}
	if gotHistoryBranchID != branchID {
		t.Fatalf("expected history branch %s, got %s", branchID, gotHistoryBranchID)
	}
}

func TestCreateProductStockSkipsHistoryWhenUnitLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	branchID := primitive.NewObjectID().Hex()
	balanceCalled := false
	historyCalled := false

	stockRepo := &productStockRepoStub{
		createStockFn: func(param request.ProductStock) (*entities.ProductStock, error) {
			return &entities.ProductStock{
				Id:        primitive.NewObjectID(),
				ProductId: productID,
				UnitId:    unitID,
			}, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int {
			balanceCalled = true
			return 0
		},
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			historyCalled = true
			return &entities.ProductHistory{}, nil
		},
	}
	productRepo := &productStockProductStub{
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return nil, errors.New("unit lookup failed")
		},
	}

	body := `{"productId":"` + productID.Hex() + `","unitId":"` + unitID.Hex() + `","quantity":2,"expireDate":"` + time.Now().UTC().Format(time.RFC3339) + `","importDate":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/product-stocks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateProductStock(stockRepo, productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if balanceCalled {
		t.Fatal("expected balance lookup to be skipped when unit lookup fails")
	}
	if historyCalled {
		t.Fatal("expected history creation to be skipped when unit lookup fails")
	}
}

func TestUpdateProductStockSequencePassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	stockID := primitive.NewObjectID().Hex()
	var gotReq request.UpdateProductStockSequence

	repo := &productStockRepoStub{
		updateSequenceFn: func(param request.UpdateProductStockSequence) ([]entities.ProductStock, error) {
			gotReq = param
			return []entities.ProductStock{}, nil
		},
	}

	body := `{"stocks":[{"stockId":"` + stockID + `","sequence":1}]}`
	req := httptest.NewRequest(http.MethodPatch, "/product-stocks/sequence", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", branchID)

	UpdateProductStockSequence(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotReq.BranchId != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotReq.BranchId)
	}
	if len(gotReq.Stocks) != 1 || gotReq.Stocks[0].StockId != stockID {
		t.Fatalf("expected stocks payload to be forwarded, got %+v", gotReq.Stocks)
	}
}

func TestUpdateProductStockByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := primitive.NewObjectID().Hex()
	unitID := primitive.NewObjectID().Hex()
	stockID := primitive.NewObjectID().Hex()
	now := time.Now().UTC().Format(time.RFC3339)

	stockRepo := &productStockRepoStub{
		getStockByIDFn: func(id string) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		updateStockFn: func(id string, param request.UpdateProductStock) (*entities.ProductStock, error) {
			t.Fatal("stock update should not run for foreign branch")
			return nil, nil
		},
	}
	productRepo := &productStockProductStub{}

	body := `{"productId":"` + productID + `","unitId":"` + unitID + `","expireDate":"` + now + `","importDate":"` + now + `"}`
	req := httptest.NewRequest(http.MethodPut, "/products/stocks/"+stockID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "stockId", Value: stockID}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())
	ctx.Set("UserId", "user-1")

	UpdateProductStockById(stockRepo, productRepo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestUpdateProductStockQuantityByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stockRepo := &productStockRepoStub{
		getStockByIDFn: func(id string) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		updateQtyFn: func(id string, quantity int) (*entities.ProductStock, error) {
			t.Fatal("stock quantity update should not run for foreign branch")
			return nil, nil
		},
	}
	productRepo := &productStockProductStub{}

	req := httptest.NewRequest(http.MethodPatch, "/products/stocks/"+primitive.NewObjectID().Hex()+"/quantity", strings.NewReader(`{"quantity":2}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "stockId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())
	ctx.Set("UserId", "user-1")

	UpdateProductStockQuantityById(stockRepo, productRepo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestRemoveProductStockByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stockRepo := &productStockRepoStub{
		getStockByIDFn: func(id string) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		removeStockFn: func(id string) (*entities.ProductStock, error) {
			t.Fatal("stock removal should not run for foreign branch")
			return nil, nil
		},
	}
	productRepo := &productStockProductStub{}

	req := httptest.NewRequest(http.MethodDelete, "/products/stocks/"+primitive.NewObjectID().Hex(), nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "stockId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())
	ctx.Set("UserId", "user-1")

	RemoveProductStockById(stockRepo, productRepo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}
