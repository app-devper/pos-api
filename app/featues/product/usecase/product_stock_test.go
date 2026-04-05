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
	repositories.IProduct
	createStockFn    func(param request.ProductStock) (*entities.ProductStock, error)
	getUnitByIDFn    func(id string) (*entities.ProductUnit, error)
	getBalanceFn     func(productId string, unitId string, branchId string) int
	createHistoryFn  func(param request.ProductHistory) (*entities.ProductHistory, error)
	updateSequenceFn func(param request.UpdateProductStockSequence) ([]entities.ProductStock, error)
}

func (s *productStockRepoStub) CreateProductStock(param request.ProductStock) (*entities.ProductStock, error) {
	return s.createStockFn(param)
}

func (s *productStockRepoStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return s.getUnitByIDFn(id)
}

func (s *productStockRepoStub) GetProductStockBalance(productId string, unitId string, branchId string) int {
	return s.getBalanceFn(productId, unitId, branchId)
}

func (s *productStockRepoStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
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

	repo := &productStockRepoStub{
		createStockFn: func(param request.ProductStock) (*entities.ProductStock, error) {
			return &entities.ProductStock{
				Id:        primitive.NewObjectID(),
				ProductId: productID,
				UnitId:    unitID,
			}, nil
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
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

	body := `{"productId":"` + productID.Hex() + `","unitId":"` + unitID.Hex() + `","quantity":2,"expireDate":"` + time.Now().UTC().Format(time.RFC3339) + `","importDate":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/product-stocks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateProductStock(repo)(ctx)

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

	repo := &productStockRepoStub{
		createStockFn: func(param request.ProductStock) (*entities.ProductStock, error) {
			return &entities.ProductStock{
				Id:        primitive.NewObjectID(),
				ProductId: productID,
				UnitId:    unitID,
			}, nil
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return nil, errors.New("unit lookup failed")
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

	body := `{"productId":"` + productID.Hex() + `","unitId":"` + unitID.Hex() + `","quantity":2,"expireDate":"` + time.Now().UTC().Format(time.RFC3339) + `","importDate":"` + time.Now().UTC().Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/product-stocks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateProductStock(repo)(ctx)

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
