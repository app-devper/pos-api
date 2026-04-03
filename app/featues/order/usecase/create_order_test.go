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
	createOrderFn   func(form request.Order) (*entities.Order, error)
	removeOrderById func(id string) (*entities.OrderDetail, error)
}

func (s *orderRepoStub) CreateOrder(form request.Order) (*entities.Order, error) {
	return s.createOrderFn(form)
}

func (s *orderRepoStub) RemoveOrderById(id string) (*entities.OrderDetail, error) {
	return s.removeOrderById(id)
}

type productStub struct {
	repositories.IProduct
	removeStockFn     func(stockId string, quantity int) (*entities.ProductStock, error)
	addStockFn        func(stockId string, quantity int) (*entities.ProductStock, error)
	getUnitByIDFn     func(id string) (*entities.ProductUnit, error)
	getBalanceFn      func(productId string, unitId string, branchId string) int
	createHistoryFn   func(param request.ProductHistory) (*entities.ProductHistory, error)
	removeHistoryFn   func(id string) (*entities.ProductHistory, error)
	removeSoldFirstFn func(productId string, quantity int) (*entities.Product, error)
	addSoldFirstFn    func(productId string, quantity int) (*entities.Product, error)
}

func (s *productStub) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.removeStockFn(stockId, quantity)
}

func (s *productStub) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.addStockFn(stockId, quantity)
}

func (s *productStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return s.getUnitByIDFn(id)
}

func (s *productStub) GetProductStockBalance(productId string, unitId string, branchId string) int {
	return s.getBalanceFn(productId, unitId, branchId)
}

func (s *productStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

func (s *productStub) RemoveProductHistoryById(id string) (*entities.ProductHistory, error) {
	return s.removeHistoryFn(id)
}

func (s *productStub) RemoveQuantitySoldFirstById(productId string, quantity int) (*entities.Product, error) {
	return s.removeSoldFirstFn(productId, quantity)
}

func (s *productStub) AddQuantitySoldFirstById(productId string, quantity int) (*entities.Product, error) {
	return s.addSoldFirstFn(productId, quantity)
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
		createOrderFn: func(form request.Order) (*entities.Order, error) {
			return &entities.Order{Id: orderID}, nil
		},
		removeOrderById: func(id string) (*entities.OrderDetail, error) {
			removedOrderID = id
			return &entities.OrderDetail{}, nil
		},
	}

	productRepo := &productStub{
		removeStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: stockID}, nil
		},
		addStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			restoredStockID = stockId
			restoredStockQty = quantity
			return &entities.ProductStock{}, nil
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int { return 5 },
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			return nil, errors.New("history failed")
		},
		removeHistoryFn:   func(id string) (*entities.ProductHistory, error) { return &entities.ProductHistory{}, nil },
		removeSoldFirstFn: func(productId string, quantity int) (*entities.Product, error) { return &entities.Product{}, nil },
		addSoldFirstFn:    func(productId string, quantity int) (*entities.Product, error) { return &entities.Product{}, nil },
	}

	sequenceRepo := &sequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) { return nil, nil },
	}

	body := `{"items":[{"productId":"` + productID.Hex() + `","quantity":2,"unitId":"` + unitID.Hex() + `","price":10,"costPrice":5,"stocks":[{"stockId":"` + stockID.Hex() + `","quantity":2}]}],"amount":20,"type":"cash","total":20}`
	req := httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateOrder(orderRepo, productRepo, sequenceRepo)(ctx)

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
