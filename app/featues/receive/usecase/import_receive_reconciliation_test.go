package usecase

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type reconcileOrderStub struct {
	repositories.IOrder
	oversoldItems []entities.OrderItem
	drainCalls    []struct {
		orderItemId string
		drain       int
		stockRef    string
	}
}

func (s *reconcileOrderStub) GetOversoldOrderItemsByProductId(productId string, branchId string) ([]entities.OrderItem, error) {
	return s.oversoldItems, nil
}

func (s *reconcileOrderStub) DrainOversoldQtyByOrderItemId(orderItemId string, drain int, stockRef string) (*entities.OrderItem, error) {
	s.drainCalls = append(s.drainCalls, struct {
		orderItemId string
		drain       int
		stockRef    string
	}{orderItemId, drain, stockRef})
	return &entities.OrderItem{}, nil
}

type reconcileStockStub struct {
	repositories.IProductStock
	lots        []entities.ProductStock
	removeCalls []struct {
		stockId  string
		quantity int
	}
}

func (s *reconcileStockStub) GetProductStocksByProductIdAndReceiveCode(productId string, receiveCode string, branchId string) ([]entities.ProductStock, error) {
	return s.lots, nil
}

func (s *reconcileStockStub) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	s.removeCalls = append(s.removeCalls, struct {
		stockId  string
		quantity int
	}{stockId, quantity})
	return &entities.ProductStock{}, nil
}

func (s *reconcileStockStub) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return &entities.ProductStock{}, nil
}

func TestImportReceiveToStockReconcilesOversoldOrdersAgainstNewLot(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	orderItemID := primitive.NewObjectID()
	lotID := primitive.NewObjectID()

	receiveRepo := &receiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID, BranchId: branchID}, nil
		},
		importReceiveToStockFn: func(receiveId string, userId string, branchId string) (*entities.Receive, error) {
			return &entities.Receive{
				Id:   receiveID,
				Code: "RC-0001",
				Items: []entities.ReceiveItem{
					{ProductId: productID, Quantity: 10},
				},
			}, nil
		},
	}

	orderRepo := &reconcileOrderStub{
		oversoldItems: []entities.OrderItem{
			{Id: orderItemID, OversoldQty: 4},
		},
	}
	stockRepo := &reconcileStockStub{
		lots: []entities.ProductStock{
			{Id: lotID, Quantity: 10},
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/receives/"+receiveID.Hex()+"/import", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID.Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	ImportReceiveToStockWithReconciliation(receiveRepo, nil, orderRepo, stockRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if len(stockRepo.removeCalls) != 1 || stockRepo.removeCalls[0].stockId != lotID.Hex() || stockRepo.removeCalls[0].quantity != 4 {
		t.Fatalf("expected lot %s to be drained by 4, got %+v", lotID.Hex(), stockRepo.removeCalls)
	}
	if len(orderRepo.drainCalls) != 1 || orderRepo.drainCalls[0].orderItemId != orderItemID.Hex() || orderRepo.drainCalls[0].drain != 4 || orderRepo.drainCalls[0].stockRef != lotID.Hex() {
		t.Fatalf("expected oversold order item %s to be drained by 4 against lot %s, got %+v", orderItemID.Hex(), lotID.Hex(), orderRepo.drainCalls)
	}
}

func TestImportReceiveToStockSkipsReconciliationWhenNoDependenciesProvided(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()

	receiveRepo := &receiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID, BranchId: branchID}, nil
		},
		importReceiveToStockFn: func(receiveId string, userId string, branchId string) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID, Code: "RC-0002"}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/receives/"+receiveID.Hex()+"/import", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID.Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	ImportReceiveToStock(receiveRepo, nil)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}
