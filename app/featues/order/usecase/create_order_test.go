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
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type orderRepoStub struct {
	repositories.IOrder
	createOrderFn         func(form request.Order) (*entities.Order, []entities.OrderItem, error)
	removeOrderById       func(id string) (*entities.OrderDetail, error)
	updateAllocationFn    func(orderItemId string, stocks []entities.OrderItemStock, oversoldQty int) (*entities.OrderItem, error)
	updateAllocationCalls []struct {
		orderItemId string
		stocks      []entities.OrderItemStock
		oversoldQty int
	}
}

func (s *orderRepoStub) CreateOrder(form request.Order) (*entities.Order, []entities.OrderItem, error) {
	return s.createOrderFn(form)
}

func (s *orderRepoStub) RemoveOrderById(id string) (*entities.OrderDetail, error) {
	return s.removeOrderById(id)
}

func (s *orderRepoStub) UpdateOrderItemAllocationById(orderItemId string, stocks []entities.OrderItemStock, oversoldQty int) (*entities.OrderItem, error) {
	s.updateAllocationCalls = append(s.updateAllocationCalls, struct {
		orderItemId string
		stocks      []entities.OrderItemStock
		oversoldQty int
	}{orderItemId, stocks, oversoldQty})
	if s.updateAllocationFn != nil {
		return s.updateAllocationFn(orderItemId, stocks, oversoldQty)
	}
	return &entities.OrderItem{OversoldQty: oversoldQty, Stocks: stocks}, nil
}

type productStub struct {
	repositories.IProduct
	getUnitByIDFn     func(id string) (*entities.ProductUnit, error)
	removeSoldFirstFn func(productId string, quantity int) (*entities.Product, error)
	addSoldFirstFn    func(productId string, quantity int) (*entities.Product, error)
}

func (s *productStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return s.getUnitByIDFn(id)
}

func (s *productStub) RemoveQuantitySoldFirstById(productId string, quantity int) (*entities.Product, error) {
	return s.removeSoldFirstFn(productId, quantity)
}

func (s *productStub) AddQuantitySoldFirstById(productId string, quantity int) (*entities.Product, error) {
	return s.addSoldFirstFn(productId, quantity)
}

type productStockStub struct {
	repositories.IProductStock
	removeStockFn   func(stockId string, quantity int) (*entities.ProductStock, error)
	addStockFn      func(stockId string, quantity int) (*entities.ProductStock, error)
	drainStockFn    func(stockId string, quantity int) (*entities.ProductStock, int, error)
	getBalanceFn    func(productId string, unitId string, branchId string) int
	createHistoryFn func(param request.ProductHistory) (*entities.ProductHistory, error)
	removeHistoryFn func(id string) (*entities.ProductHistory, error)
}

func (s *productStockStub) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.removeStockFn(stockId, quantity)
}

func (s *productStockStub) DrainProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, int, error) {
	return s.drainStockFn(stockId, quantity)
}

func (s *productStockStub) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.addStockFn(stockId, quantity)
}

func (s *productStockStub) GetProductStockBalance(productId string, unitId string, branchId string) int {
	return s.getBalanceFn(productId, unitId, branchId)
}

func (s *productStockStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

func (s *productStockStub) RemoveProductHistoryById(id string) (*entities.ProductHistory, error) {
	return s.removeHistoryFn(id)
}

type sequenceStub struct {
	repositories.ISequence
	nextFn func(field string) (*entities.Sequence, error)
}

func (s *sequenceStub) NextSequence(field string) (*entities.Sequence, error) {
	return s.nextFn(field)
}

func TestCreateOrderRollsBackWhenHistoryCreationFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	stockID := primitive.NewObjectID()
	productID := primitive.NewObjectID()

	var removedOrderID string
	var restoredStockID string
	var restoredStockQty int

	orderRepo := &orderRepoStub{
		createOrderFn: func(form request.Order) (*entities.Order, []entities.OrderItem, error) {
			return &entities.Order{Id: orderID}, nil, nil
		},
		removeOrderById: func(id string) (*entities.OrderDetail, error) {
			removedOrderID = id
			return &entities.OrderDetail{}, nil
		},
	}

	productRepo := &productStub{
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
		},
		removeSoldFirstFn: func(productId string, quantity int) (*entities.Product, error) { return &entities.Product{}, nil },
		addSoldFirstFn:    func(productId string, quantity int) (*entities.Product, error) { return &entities.Product{}, nil },
	}
	productStockRepo := &productStockStub{
		removeStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: stockID}, nil
		},
		addStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			restoredStockID = stockId
			restoredStockQty = quantity
			return &entities.ProductStock{}, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int { return 5 },
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			return nil, errors.New("history failed")
		},
		removeHistoryFn: func(id string) (*entities.ProductHistory, error) { return &entities.ProductHistory{}, nil },
	}

	sequenceRepo := &sequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: field, Prefix: "OR", Value: 1, Format: 4}, nil
		},
	}

	body := `{"items":[{"productId":"` + productID.Hex() + `","quantity":2,"unitId":"` + unitID.Hex() + `","price":10,"costPrice":5,"stocks":[{"stockId":"` + stockID.Hex() + `","quantity":2}]}],"amount":20,"type":"cash","total":20}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateOrder(orderRepo, productRepo, productStockRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if removedOrderID != orderID.Hex() {
		t.Fatalf("expected order rollback for %s, got %s", orderID.Hex(), removedOrderID)
	}
	if restoredStockID != stockID.Hex() || restoredStockQty != 2 {
		t.Fatalf("expected stock rollback for %s qty 2, got %s qty %d", stockID.Hex(), restoredStockID, restoredStockQty)
	}
	if !strings.Contains(w.Body.String(), errcode.OR_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.OR_BAD_REQUEST_002, w.Body.String())
	}
}

func TestCreateOrderFailsWhenSequenceLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := &orderRepoStub{
		createOrderFn: func(form request.Order) (*entities.Order, []entities.OrderItem, error) {
			t.Fatal("create order should not be called when sequence fails")
			return nil, nil, nil
		},
		removeOrderById: func(id string) (*entities.OrderDetail, error) {
			return &entities.OrderDetail{}, nil
		},
	}
	productRepo := &productStub{}
	productStockRepo := &productStockStub{}
	sequenceRepo := &sequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return nil, errors.New("sequence failed")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[],"amount":20,"type":"cash","total":20}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateOrder(orderRepo, productRepo, productStockRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "sequence failed") {
		t.Fatalf("expected sequence failure in response, got %s", w.Body.String())
	}
}

func TestCreateOrderFailsWhenSequenceIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := &orderRepoStub{
		createOrderFn: func(form request.Order) (*entities.Order, []entities.OrderItem, error) {
			t.Fatal("create order should not be called when sequence is missing")
			return nil, nil, nil
		},
		removeOrderById: func(id string) (*entities.OrderDetail, error) {
			return &entities.OrderDetail{}, nil
		},
	}
	productRepo := &productStub{}
	productStockRepo := &productStockStub{}
	sequenceRepo := &sequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) { return nil, nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(`{"items":[],"amount":20,"type":"cash","total":20}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateOrder(orderRepo, productRepo, productStockRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "order sequence not available") {
		t.Fatalf("expected missing sequence error, got %s", w.Body.String())
	}
}

func TestCreateOrderRejectsInsufficientStockWhenOversellNotAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	stockID := primitive.NewObjectID()
	productID := primitive.NewObjectID()

	var removedOrderID string

	orderRepo := &orderRepoStub{
		createOrderFn: func(form request.Order) (*entities.Order, []entities.OrderItem, error) {
			return &entities.Order{Id: orderID}, []entities.OrderItem{{Id: primitive.NewObjectID()}}, nil
		},
		removeOrderById: func(id string) (*entities.OrderDetail, error) {
			removedOrderID = id
			return &entities.OrderDetail{}, nil
		},
	}
	productRepo := &productStub{}
	productStockRepo := &productStockStub{
		removeStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return nil, errors.New("quantity not available")
		},
	}
	sequenceRepo := &sequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: field, Prefix: "OR", Value: 1, Format: 4}, nil
		},
	}

	body := `{"items":[{"productId":"` + productID.Hex() + `","quantity":5,"unitId":"` + unitID.Hex() + `","price":10,"costPrice":5,"stocks":[{"stockId":"` + stockID.Hex() + `","quantity":5}]}],"amount":50,"type":"cash","total":50}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateOrder(orderRepo, productRepo, productStockRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if removedOrderID != orderID.Hex() {
		t.Fatalf("expected order rollback for %s, got %s", orderID.Hex(), removedOrderID)
	}
	if len(orderRepo.updateAllocationCalls) != 0 {
		t.Fatalf("expected no oversold allocation recorded, got %+v", orderRepo.updateAllocationCalls)
	}
}

func TestCreateOrderRecordsOversoldQtyWhenAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID()
	orderItemID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	stockID := primitive.NewObjectID()
	productID := primitive.NewObjectID()

	orderRepo := &orderRepoStub{
		createOrderFn: func(form request.Order) (*entities.Order, []entities.OrderItem, error) {
			return &entities.Order{Id: orderID}, []entities.OrderItem{{Id: orderItemID}}, nil
		},
		removeOrderById: func(id string) (*entities.OrderDetail, error) {
			return &entities.OrderDetail{}, nil
		},
	}
	productRepo := &productStub{
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
		},
	}
	productStockRepo := &productStockStub{
		removeStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return nil, errors.New("quantity not available")
		},
		drainStockFn: func(stockId string, quantity int) (*entities.ProductStock, int, error) {
			return &entities.ProductStock{Id: stockID, Quantity: 0}, 3, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int { return 0 },
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			return &entities.ProductHistory{Id: primitive.NewObjectID()}, nil
		},
	}
	sequenceRepo := &sequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: field, Prefix: "OR", Value: 1, Format: 4}, nil
		},
	}

	body := `{"items":[{"productId":"` + productID.Hex() + `","quantity":5,"unitId":"` + unitID.Hex() + `","price":10,"costPrice":5,"allowOversell":true,"stocks":[{"stockId":"` + stockID.Hex() + `","quantity":5}]}],"amount":50,"type":"cash","total":50}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateOrder(orderRepo, productRepo, productStockRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if len(orderRepo.updateAllocationCalls) != 1 {
		t.Fatalf("expected exactly one oversold allocation recorded, got %+v", orderRepo.updateAllocationCalls)
	}
	call := orderRepo.updateAllocationCalls[0]
	if call.orderItemId != orderItemID.Hex() {
		t.Fatalf("expected allocation recorded for order item %s, got %s", orderItemID.Hex(), call.orderItemId)
	}
	if call.oversoldQty != 2 {
		t.Fatalf("expected oversoldQty 2 (5 requested - 3 drained), got %d", call.oversoldQty)
	}
	if len(call.stocks) != 1 || call.stocks[0].Quantity != 3 || call.stocks[0].StockId != stockID.Hex() {
		t.Fatalf("expected finalized stock allocation to reflect drained qty 3, got %+v", call.stocks)
	}
}
