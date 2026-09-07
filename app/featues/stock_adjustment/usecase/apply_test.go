package usecase

import (
	"errors"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type adjustmentRepoStub struct {
	repositories.IStockAdjustment
	createFn func(param repositories.StockAdjustmentInput) (*entities.StockAdjustment, error)
}

func (s *adjustmentRepoStub) CreateStockAdjustment(param repositories.StockAdjustmentInput) (*entities.StockAdjustment, error) {
	return s.createFn(param)
}

type adjustmentProductStockStub struct {
	repositories.IProductStock
	getStockByIdFn  func(id string) (*entities.ProductStock, error)
	addStockFn      func(stockId string, quantity int) (*entities.ProductStock, error)
	removeStockFn   func(stockId string, quantity int) (*entities.ProductStock, error)
	getBalanceFn    func(productId string, unitId string, branchId string) int
	createHistoryFn func(param request.ProductHistory) (*entities.ProductHistory, error)
}

func (s *adjustmentProductStockStub) GetProductStockById(id string) (*entities.ProductStock, error) {
	return s.getStockByIdFn(id)
}
func (s *adjustmentProductStockStub) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.addStockFn(stockId, quantity)
}
func (s *adjustmentProductStockStub) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.removeStockFn(stockId, quantity)
}
func (s *adjustmentProductStockStub) GetProductStockBalance(productId string, unitId string, branchId string) int {
	return s.getBalanceFn(productId, unitId, branchId)
}
func (s *adjustmentProductStockStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

type adjustmentProductStub struct {
	repositories.IProduct
	getUnitByIdFn func(id string) (*entities.ProductUnit, error)
}

func (s *adjustmentProductStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return s.getUnitByIdFn(id)
}

type adjustmentOrderStub struct {
	repositories.IOrder
	oversoldItems []entities.OrderItem
	drainCalls    []struct {
		orderItemId string
		drain       int
		stockRef    string
	}
}

func (s *adjustmentOrderStub) GetOversoldOrderItemsByProductId(productId string, branchId string) ([]entities.OrderItem, error) {
	return s.oversoldItems, nil
}

func (s *adjustmentOrderStub) DrainOversoldQtyByOrderItemId(orderItemId string, drain int, stockRef string) (*entities.OrderItem, error) {
	s.drainCalls = append(s.drainCalls, struct {
		orderItemId string
		drain       int
		stockRef    string
	}{orderItemId, drain, stockRef})
	return &entities.OrderItem{}, nil
}

type adjustmentSequenceStub struct {
	repositories.ISequence
	nextFn func(field string) (*entities.Sequence, error)
}

func (s *adjustmentSequenceStub) NextSequence(field string) (*entities.Sequence, error) {
	return s.nextFn(field)
}

func TestApplyAdjustmentRejectsZeroDelta(t *testing.T) {
	_, err := ApplyAdjustment(&adjustmentRepoStub{}, &adjustmentProductStockStub{}, &adjustmentProductStub{}, &adjustmentOrderStub{}, &adjustmentSequenceStub{}, request.StockAdjustment{Delta: 0, Reason: constant.AdjustmentReasonOther})
	if err == nil {
		t.Fatal("expected error for zero delta, got nil")
	}
}

func TestApplyAdjustmentRejectsInvalidReason(t *testing.T) {
	_, err := ApplyAdjustment(&adjustmentRepoStub{}, &adjustmentProductStockStub{}, &adjustmentProductStub{}, &adjustmentOrderStub{}, &adjustmentSequenceStub{}, request.StockAdjustment{Delta: 5, Reason: "not-a-real-reason"})
	if err == nil {
		t.Fatal("expected error for invalid reason, got nil")
	}
}

func TestApplyAdjustmentIncreasesStockAndReconcilesOversold(t *testing.T) {
	stockID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	orderItemID := primitive.NewObjectID()

	adjustmentRepo := &adjustmentRepoStub{
		createFn: func(param repositories.StockAdjustmentInput) (*entities.StockAdjustment, error) {
			return &entities.StockAdjustment{Id: primitive.NewObjectID(), Delta: param.Delta, Before: param.Before, After: param.After}, nil
		},
	}
	productStockRepo := &adjustmentProductStockStub{
		getStockByIdFn: func(id string) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: stockID, BranchId: branchID, UnitId: unitID, Quantity: 3}, nil
		},
		addStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: stockID, Quantity: 13}, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int { return 13 },
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			return &entities.ProductHistory{Id: primitive.NewObjectID()}, nil
		},
	}
	productRepo := &adjustmentProductStub{
		getUnitByIdFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
		},
	}
	orderRepo := &adjustmentOrderStub{
		oversoldItems: []entities.OrderItem{{Id: orderItemID, OversoldQty: 6}},
	}
	sequenceRepo := &adjustmentSequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: field, Prefix: "", Value: 1, Format: 4}, nil
		},
	}

	req := request.StockAdjustment{
		ProductId: productID.Hex(),
		StockId:   stockID.Hex(),
		Reason:    constant.AdjustmentReasonCount,
		Delta:     10,
		BranchId:  branchID.Hex(),
		CreatedBy: "user-1",
	}

	result, err := ApplyAdjustment(adjustmentRepo, productStockRepo, productRepo, orderRepo, sequenceRepo, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Before != 3 || result.After != 13 || result.Delta != 10 {
		t.Fatalf("expected before=3 after=13 delta=10, got %+v", result)
	}
	if len(orderRepo.drainCalls) != 1 {
		t.Fatalf("expected one oversold drain call, got %+v", orderRepo.drainCalls)
	}
	call := orderRepo.drainCalls[0]
	if call.orderItemId != orderItemID.Hex() || call.drain != 6 || call.stockRef != "ADJUST:"+constant.AdjustmentReasonCount {
		t.Fatalf("unexpected drain call: %+v", call)
	}
}

func TestApplyAdjustmentDecreaseDoesNotReconcileOversold(t *testing.T) {
	stockID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()

	adjustmentRepo := &adjustmentRepoStub{
		createFn: func(param repositories.StockAdjustmentInput) (*entities.StockAdjustment, error) {
			return &entities.StockAdjustment{Id: primitive.NewObjectID(), Delta: param.Delta, Before: param.Before, After: param.After}, nil
		},
	}
	productStockRepo := &adjustmentProductStockStub{
		getStockByIdFn: func(id string) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: stockID, BranchId: branchID, UnitId: unitID, Quantity: 10}, nil
		},
		removeStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: stockID, Quantity: 6}, nil
		},
		getBalanceFn: func(productId string, unitId string, branchId string) int { return 6 },
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			return &entities.ProductHistory{Id: primitive.NewObjectID()}, nil
		},
	}
	productRepo := &adjustmentProductStub{
		getUnitByIdFn: func(id string) (*entities.ProductUnit, error) {
			return &entities.ProductUnit{Id: unitID, Unit: "TAB"}, nil
		},
	}
	orderRepo := &adjustmentOrderStub{}
	sequenceRepo := &adjustmentSequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: field, Prefix: "", Value: 2, Format: 4}, nil
		},
	}

	req := request.StockAdjustment{
		ProductId: productID.Hex(),
		StockId:   stockID.Hex(),
		Reason:    constant.AdjustmentReasonDamaged,
		Delta:     -4,
		BranchId:  branchID.Hex(),
		CreatedBy: "user-1",
	}

	result, err := ApplyAdjustment(adjustmentRepo, productStockRepo, productRepo, orderRepo, sequenceRepo, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Before != 10 || result.After != 6 || result.Delta != -4 {
		t.Fatalf("expected before=10 after=6 delta=-4, got %+v", result)
	}
	if len(orderRepo.drainCalls) != 0 {
		t.Fatalf("expected no oversold reconciliation for a negative adjustment, got %+v", orderRepo.drainCalls)
	}
}

func TestApplyAdjustmentFailsWhenRemovingMoreThanAvailable(t *testing.T) {
	stockID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()

	productStockRepo := &adjustmentProductStockStub{
		getStockByIdFn: func(id string) (*entities.ProductStock, error) {
			return &entities.ProductStock{Id: stockID, BranchId: branchID, Quantity: 2}, nil
		},
		removeStockFn: func(stockId string, quantity int) (*entities.ProductStock, error) {
			return nil, errors.New("quantity not available")
		},
	}

	req := request.StockAdjustment{
		ProductId: productID.Hex(),
		StockId:   stockID.Hex(),
		Reason:    constant.AdjustmentReasonLost,
		Delta:     -5,
		BranchId:  branchID.Hex(),
	}

	_, err := ApplyAdjustment(&adjustmentRepoStub{}, productStockRepo, &adjustmentProductStub{}, &adjustmentOrderStub{}, &adjustmentSequenceStub{}, req)
	if err == nil {
		t.Fatal("expected error when removing more stock than available")
	}
}
