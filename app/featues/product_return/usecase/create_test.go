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

type returnOrderStub struct {
	repositories.IOrder
	order          *entities.Order
	items          map[string]*entities.OrderItem
	incrementCalls []struct {
		orderItemId string
		quantity    int
	}
}

func (s *returnOrderStub) GetOrderById(id string) (*entities.Order, error) {
	return s.order, nil
}

func (s *returnOrderStub) GetOrderItemById(id string) (*entities.OrderItem, error) {
	return s.items[id], nil
}

func (s *returnOrderStub) IncrementOrderItemReturnedQtyById(orderItemId string, quantity int) (*entities.OrderItem, error) {
	s.incrementCalls = append(s.incrementCalls, struct {
		orderItemId string
		quantity    int
	}{orderItemId, quantity})
	return &entities.OrderItem{}, nil
}

type returnProductStockStub struct {
	repositories.IProductStock
	addCalls []struct {
		stockId  string
		quantity int
	}
}

func (s *returnProductStockStub) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	s.addCalls = append(s.addCalls, struct {
		stockId  string
		quantity int
	}{stockId, quantity})
	return &entities.ProductStock{}, nil
}

func (s *returnProductStockStub) GetProductStockBalance(productId string, unitId string, branchId string) int {
	return 0
}

func (s *returnProductStockStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return &entities.ProductHistory{Id: primitive.NewObjectID()}, nil
}

type returnProductStub struct {
	repositories.IProduct
}

func (s *returnProductStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return &entities.ProductUnit{Id: primitive.NewObjectID(), Unit: "TAB"}, nil
}

type returnReturnRepoStub struct {
	repositories.IProductReturn
	createFn func(param repositories.ProductReturnInput) (*entities.ProductReturn, error)
}

func (s *returnReturnRepoStub) CreateProductReturn(param repositories.ProductReturnInput) (*entities.ProductReturn, error) {
	return s.createFn(param)
}

type returnSequenceStub struct {
	repositories.ISequence
}

func (s *returnSequenceStub) NextSequence(field string) (*entities.Sequence, error) {
	return &entities.Sequence{Field: field, Prefix: "", Value: 1, Format: 4}, nil
}

func TestCreateProductReturnRestoresStockAndRecordsReturn(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID()
	orderItemID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	lotID := primitive.NewObjectID()

	orderRepo := &returnOrderStub{
		order: &entities.Order{Id: orderID, BranchId: branchID, CustomerCode: "C001"},
		items: map[string]*entities.OrderItem{
			orderItemID.Hex(): {
				Id: orderItemID, OrderId: orderID, ProductId: productID, UnitId: unitID,
				Quantity: 5, ReturnedQty: 0, Price: 20,
				Stocks: []entities.OrderItemStock{{StockId: lotID.Hex(), Quantity: 5}},
			},
		},
	}
	productStockRepo := &returnProductStockStub{}
	productRepo := &returnProductStub{}
	var createdInput repositories.ProductReturnInput
	returnRepo := &returnReturnRepoStub{
		createFn: func(param repositories.ProductReturnInput) (*entities.ProductReturn, error) {
			createdInput = param
			return &entities.ProductReturn{Id: primitive.NewObjectID(), ReturnNo: param.ReturnNo, Items: param.Items, TotalRefund: param.TotalRefund}, nil
		},
	}
	sequenceRepo := &returnSequenceStub{}

	body := `{"orderId":"` + orderID.Hex() + `","reason":"ลูกค้าคืนสินค้า","items":[{"orderItemId":"` + orderItemID.Hex() + `","quantity":2,"refund":40}]}`
	req := httptest.NewRequest(http.MethodPost, "/product-returns", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	CreateProductReturn(returnRepo, orderRepo, productStockRepo, productRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if len(productStockRepo.addCalls) != 1 || productStockRepo.addCalls[0].stockId != lotID.Hex() || productStockRepo.addCalls[0].quantity != 2 {
		t.Fatalf("expected lot %s to be restored by 2, got %+v", lotID.Hex(), productStockRepo.addCalls)
	}
	if len(orderRepo.incrementCalls) != 1 || orderRepo.incrementCalls[0].orderItemId != orderItemID.Hex() || orderRepo.incrementCalls[0].quantity != 2 {
		t.Fatalf("expected returnedQty increment of 2 on order item %s, got %+v", orderItemID.Hex(), orderRepo.incrementCalls)
	}
	if createdInput.TotalRefund != 40 || createdInput.CustomerCode != "C001" {
		t.Fatalf("expected total refund 40 and customer code C001, got %+v", createdInput)
	}
}

func TestCreateProductReturnRejectsReturnBeyondRealLotQuantity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderID := primitive.NewObjectID()
	orderItemID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	lotID := primitive.NewObjectID()

	orderRepo := &returnOrderStub{
		order: &entities.Order{Id: orderID, BranchId: branchID},
		items: map[string]*entities.OrderItem{
			orderItemID.Hex(): {
				Id: orderItemID, OrderId: orderID, ProductId: productID, UnitId: unitID,
				Quantity: 5, ReturnedQty: 0, Price: 20,
				Stocks: []entities.OrderItemStock{
					{StockId: lotID.Hex(), Quantity: 3},
					{StockId: "ADJUST:สูญหาย", Quantity: 2},
				},
			},
		},
	}
	productStockRepo := &returnProductStockStub{}
	productRepo := &returnProductStub{}
	returnRepo := &returnReturnRepoStub{
		createFn: func(param repositories.ProductReturnInput) (*entities.ProductReturn, error) {
			t.Fatal("CreateProductReturn should not be called when the requested quantity exceeds real lot backing")
			return nil, nil
		},
	}
	sequenceRepo := &returnSequenceStub{}

	body := `{"orderId":"` + orderID.Hex() + `","reason":"ลูกค้าคืนสินค้า","items":[{"orderItemId":"` + orderItemID.Hex() + `","quantity":4,"refund":80}]}`
	req := httptest.NewRequest(http.MethodPost, "/product-returns", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	CreateProductReturn(returnRepo, orderRepo, productStockRepo, productRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.RT_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.RT_BAD_REQUEST_002, w.Body.String())
	}
	if len(productStockRepo.addCalls) != 0 {
		t.Fatalf("expected no stock mutation on rejected return, got %+v", productStockRepo.addCalls)
	}
}
