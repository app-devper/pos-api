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

type countRepoStub struct {
	repositories.IStockCount
	createFn func(param repositories.StockCountInput) (*entities.StockCount, error)
}

func (s *countRepoStub) CreateStockCount(param repositories.StockCountInput) (*entities.StockCount, error) {
	return s.createFn(param)
}

type countAdjustmentRepoStub struct {
	repositories.IStockAdjustment
	createFn func(param repositories.StockAdjustmentInput) (*entities.StockAdjustment, error)
}

func (s *countAdjustmentRepoStub) CreateStockAdjustment(param repositories.StockAdjustmentInput) (*entities.StockAdjustment, error) {
	return s.createFn(param)
}

type countProductStockStub struct {
	repositories.IProductStock
	stocks          map[string]*entities.ProductStock
	addStockFn      func(stockId string, quantity int) (*entities.ProductStock, error)
	removeStockFn   func(stockId string, quantity int) (*entities.ProductStock, error)
	getBalanceFn    func(productId string, unitId string, branchId string) int
	createHistoryFn func(param request.ProductHistory) (*entities.ProductHistory, error)
}

func (s *countProductStockStub) GetProductStockById(id string) (*entities.ProductStock, error) {
	return s.stocks[id], nil
}
func (s *countProductStockStub) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.addStockFn(stockId, quantity)
}
func (s *countProductStockStub) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.removeStockFn(stockId, quantity)
}
func (s *countProductStockStub) GetProductStockBalance(productId string, unitId string, branchId string) int {
	return s.getBalanceFn(productId, unitId, branchId)
}
func (s *countProductStockStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

type countProductStub struct {
	repositories.IProduct
	getUnitByIdFn func(id string) (*entities.ProductUnit, error)
}

func (s *countProductStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return s.getUnitByIdFn(id)
}

type countOrderStub struct {
	repositories.IOrder
}

func (s *countOrderStub) GetOversoldOrderItemsByProductId(productId string, branchId string) ([]entities.OrderItem, error) {
	return nil, nil
}

type countSequenceStub struct {
	repositories.ISequence
	value int
}

func (s *countSequenceStub) NextSequence(field string) (*entities.Sequence, error) {
	s.value++
	return &entities.Sequence{Field: field, Prefix: "", Value: s.value, Format: 4}, nil
}

func TestCreateStockCountAppliesAdjustmentOnlyWhenDeltaNonZero(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	matchingStockID := primitive.NewObjectID()
	mismatchStockID := primitive.NewObjectID()

	var adjustmentCalls int

	countRepo := &countRepoStub{
		createFn: func(param repositories.StockCountInput) (*entities.StockCount, error) {
			return &entities.StockCount{Id: primitive.NewObjectID(), CountNo: param.CountNo, Items: param.Items}, nil
		},
	}
	adjustmentRepo := &countAdjustmentRepoStub{
		createFn: func(param repositories.StockAdjustmentInput) (*entities.StockAdjustment, error) {
			adjustmentCalls++
			return &entities.StockAdjustment{Id: primitive.NewObjectID()}, nil
		},
	}
	productStockRepo := &countProductStockStub{
		stocks: map[string]*entities.ProductStock{
			matchingStockID.Hex(): {Id: matchingStockID, BranchId: branchID, UnitId: unitID, Quantity: 5},
			mismatchStockID.Hex(): {Id: mismatchStockID, BranchId: branchID, UnitId: unitID, Quantity: 5},
		},
		addStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: mismatchStockID, Quantity: 5 + quantity}, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int { return 8 },
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			return &entities.ProductHistory{Id: primitive.NewObjectID()}, nil
		},
	}
	productRepo := &countProductStub{
		getUnitByIdFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
		},
	}
	orderRepo := &countOrderStub{}
	sequenceRepo := &countSequenceStub{}

	body := `{"note":"monthly count","items":[
		{"productId":"` + productID.Hex() + `","stockId":"` + matchingStockID.Hex() + `","counted":5},
		{"productId":"` + productID.Hex() + `","stockId":"` + mismatchStockID.Hex() + `","counted":8}
	]}`
	req := httptest.NewRequest(http.MethodPost, "/stock-counts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	CreateStockCount(countRepo, adjustmentRepo, productStockRepo, productRepo, orderRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if adjustmentCalls != 1 {
		t.Fatalf("expected exactly one adjustment for the mismatched line, got %d", adjustmentCalls)
	}
}
