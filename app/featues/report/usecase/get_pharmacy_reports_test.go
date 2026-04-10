package usecase

import (
	"encoding/json"
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

type khy9ReceiveStub struct {
	repositories.IReceive
	getReceivesFn          func(form request.GetReceiveRange) ([]entities.Receive, error)
	getReceiveItemsByIDFn  func(receiveId string) ([]entities.ReceiveItem, error)
	getReceiveItemsByIDsFn func(receiveIds []string) ([]entities.ReceiveItem, error)
	singleItemCalls        int
	batchItemCalls         int
}

func (s *khy9ReceiveStub) GetReceives(form request.GetReceiveRange) ([]entities.Receive, error) {
	return s.getReceivesFn(form)
}

func (s *khy9ReceiveStub) GetReceiveItemsByReceiveId(receiveId string) ([]entities.ReceiveItem, error) {
	s.singleItemCalls++
	return s.getReceiveItemsByIDFn(receiveId)
}

func (s *khy9ReceiveStub) GetReceiveItemsByReceiveIds(receiveIds []string) ([]entities.ReceiveItem, error) {
	s.batchItemCalls++
	return s.getReceiveItemsByIDsFn(receiveIds)
}

type khy9ProductStub struct {
	repositories.IProduct
	getProductsByIDsFn func(ids []string) ([]entities.Product, error)
	getUnitByIDFn      func(id string) (*entities.ProductUnit, error)
	getUnitsByIDsFn    func(ids []string) ([]entities.ProductUnit, error)
	singleUnitCalls    int
	batchUnitsCalls    int
}

func (s *khy9ProductStub) GetProductsByIds(ids []string) ([]entities.Product, error) {
	return s.getProductsByIDsFn(ids)
}

func (s *khy9ProductStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	s.singleUnitCalls++
	return s.getUnitByIDFn(id)
}

func (s *khy9ProductStub) GetProductUnitsByIds(ids []string) ([]entities.ProductUnit, error) {
	s.batchUnitsCalls++
	return s.getUnitsByIDsFn(ids)
}

type khy9SupplierStub struct {
	repositories.ISupplier
	getSupplierByIDFn   func(id string) (*entities.Supplier, error)
	getSuppliersByIDsFn func(ids []string) ([]entities.Supplier, error)
	singleCalls         int
	batchCalls          int
}

func (s *khy9SupplierStub) GetSupplierById(id string) (*entities.Supplier, error) {
	s.singleCalls++
	return s.getSupplierByIDFn(id)
}

func (s *khy9SupplierStub) GetSuppliersByIds(ids []string) ([]entities.Supplier, error) {
	s.batchCalls++
	return s.getSuppliersByIDsFn(ids)
}

type salesOrderStub struct {
	repositories.IOrder
	getOrderRangeFn     func(form request.GetOrderRange) ([]entities.Order, error)
	getOrderItemRangeFn func(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error)
}

func (s *salesOrderStub) GetOrderRange(form request.GetOrderRange) ([]entities.Order, error) {
	return s.getOrderRangeFn(form)
}

func (s *salesOrderStub) GetOrderItemRange(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error) {
	return s.getOrderItemRangeFn(form)
}

func TestGetKHY9DataUsesReceiveItemsCollection(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	supplierID := primitive.NewObjectID()
	branchID := primitive.NewObjectID().Hex()

	receiveRepo := &khy9ReceiveStub{
		getReceivesFn: func(form request.GetReceiveRange) ([]entities.Receive, error) {
			return []entities.Receive{
				{
					Id:          receiveID,
					Code:        "RC-001",
					SupplierId:  supplierID,
					CreatedDate: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC),
					Items:       []entities.ReceiveItem{},
				},
			}, nil
		},
		getReceiveItemsByIDFn: func(id string) ([]entities.ReceiveItem, error) {
			return nil, errors.New("single receive item lookup should not be used")
		},
		getReceiveItemsByIDsFn: func(ids []string) ([]entities.ReceiveItem, error) {
			return []entities.ReceiveItem{
				{
					ReceiveId:  receiveID,
					ProductId:  productID,
					UnitId:     primitive.NewObjectID().Hex(),
					Quantity:   3,
					CostPrice:  12.5,
					LotNumber:  "LOT-1",
					ExpireDate: time.Date(2027, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	}

	productRepo := &khy9ProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return []entities.Product{
				{
					Id:                productID,
					Name:              "Drug A",
					Unit:              "TAB",
					DrugRegistrations: []string{"KHY9"},
					DrugInfo:          &entities.DrugInfo{GenericName: "Gen A"},
				},
			}, nil
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return nil, errors.New("single unit lookup should not be used")
		},
		getUnitsByIDsFn: func(ids []string) ([]entities.ProductUnit, error) {
			unitID, _ := primitive.ObjectIDFromHex(ids[0])
			return []entities.ProductUnit{{Id: unitID, Unit: "BOX"}}, nil
		},
	}

	supplierRepo := &khy9SupplierStub{
		getSupplierByIDFn: func(id string) (*entities.Supplier, error) {
			return nil, errors.New("single supplier lookup should not be used")
		},
		getSuppliersByIDsFn: func(ids []string) ([]entities.Supplier, error) {
			return []entities.Supplier{{Id: supplierID, Name: "Supplier A"}}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/pharmacy/khy9/data?startDate=2026-01-01T00:00:00Z&endDate=2026-02-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", branchID)

	GetKHY9Data(receiveRepo, productRepo, supplierRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp pharmacyReportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 report item, got %d", len(resp.Items))
	}
	if resp.Items[0].ProductName != "Drug A" || resp.Items[0].Quantity != 3 {
		t.Fatalf("unexpected report item: %+v", resp.Items[0])
	}
	if resp.Items[0].Unit != "BOX" {
		t.Fatalf("expected unit BOX from receive item unitId, got %s", resp.Items[0].Unit)
	}
	if receiveRepo.batchItemCalls != 1 || receiveRepo.singleItemCalls != 0 {
		t.Fatalf("expected batch receive-items lookup once and no single lookups, got batch=%d single=%d", receiveRepo.batchItemCalls, receiveRepo.singleItemCalls)
	}
	if productRepo.batchUnitsCalls != 1 || productRepo.singleUnitCalls != 0 {
		t.Fatalf("expected batch product-unit lookup once and no single lookups, got batch=%d single=%d", productRepo.batchUnitsCalls, productRepo.singleUnitCalls)
	}
	if supplierRepo.batchCalls != 1 || supplierRepo.singleCalls != 0 {
		t.Fatalf("expected batch supplier lookup once and no single lookups, got batch=%d single=%d", supplierRepo.batchCalls, supplierRepo.singleCalls)
	}
}

func TestGetKHY9DataFailsWhenProductLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	supplierID := primitive.NewObjectID()

	receiveRepo := &khy9ReceiveStub{
		getReceivesFn: func(form request.GetReceiveRange) ([]entities.Receive, error) {
			return []entities.Receive{{Id: receiveID, SupplierId: supplierID}}, nil
		},
		getReceiveItemsByIDFn: func(id string) ([]entities.ReceiveItem, error) {
			return nil, errors.New("single receive item lookup should not be used")
		},
		getReceiveItemsByIDsFn: func(ids []string) ([]entities.ReceiveItem, error) {
			return []entities.ReceiveItem{{ReceiveId: receiveID, ProductId: productID}}, nil
		},
	}

	productRepo := &khy9ProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return nil, errors.New("lookup failed")
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return nil, errors.New("single unit lookup should not be used")
		},
		getUnitsByIDsFn: func(ids []string) ([]entities.ProductUnit, error) {
			return []entities.ProductUnit{}, nil
		},
	}

	supplierRepo := &khy9SupplierStub{
		getSupplierByIDFn: func(id string) (*entities.Supplier, error) {
			return nil, errors.New("single supplier lookup should not be used")
		},
		getSuppliersByIDsFn: func(ids []string) ([]entities.Supplier, error) {
			return []entities.Supplier{{Id: supplierID, Name: "Supplier A"}}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/pharmacy/khy9/data?startDate=2026-01-01T00:00:00Z&endDate=2026-02-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetKHY9Data(receiveRepo, productRepo, supplierRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "lookup failed") {
		t.Fatalf("expected lookup failure in response, got %s", w.Body.String())
	}
}

func TestGetKHY9DataFailsWhenSupplierLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	supplierID := primitive.NewObjectID()

	receiveRepo := &khy9ReceiveStub{
		getReceivesFn: func(form request.GetReceiveRange) ([]entities.Receive, error) {
			return []entities.Receive{{Id: receiveID, SupplierId: supplierID}}, nil
		},
		getReceiveItemsByIDFn: func(id string) ([]entities.ReceiveItem, error) {
			return nil, errors.New("single receive item lookup should not be used")
		},
		getReceiveItemsByIDsFn: func(ids []string) ([]entities.ReceiveItem, error) {
			return []entities.ReceiveItem{{ReceiveId: receiveID, ProductId: productID}}, nil
		},
	}

	productRepo := &khy9ProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return []entities.Product{{Id: productID, Name: "Drug A", DrugRegistrations: []string{"KHY9"}}}, nil
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return nil, errors.New("single unit lookup should not be used")
		},
		getUnitsByIDsFn: func(ids []string) ([]entities.ProductUnit, error) {
			return []entities.ProductUnit{}, nil
		},
	}

	supplierRepo := &khy9SupplierStub{
		getSupplierByIDFn: func(id string) (*entities.Supplier, error) {
			return nil, errors.New("single supplier lookup should not be used")
		},
		getSuppliersByIDsFn: func(ids []string) ([]entities.Supplier, error) {
			return nil, errors.New("supplier lookup failed")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/pharmacy/khy9/data?startDate=2026-01-01T00:00:00Z&endDate=2026-02-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetKHY9Data(receiveRepo, productRepo, supplierRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "supplier lookup failed") {
		t.Fatalf("expected supplier lookup failure in response, got %s", w.Body.String())
	}
}

func TestGetSalesReportItemsUsesBatchUnitLookup(t *testing.T) {
	orderID := primitive.NewObjectID()
	unitID := primitive.NewObjectID()
	branchID := primitive.NewObjectID().Hex()
	req := pharmacyReportRange{
		StartDate: request.NewFlexibleTime(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:   request.NewFlexibleTime(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)),
	}

	orderRepo := &salesOrderStub{
		getOrderRangeFn: func(form request.GetOrderRange) ([]entities.Order, error) {
			return []entities.Order{{
				Id:             orderID,
				Status:         "ACTIVE",
				PharmacistName: "เภสัชกร A",
				LicenseNo:      "1234",
			}}, nil
		},
		getOrderItemRangeFn: func(form request.GetOrderRange) ([]entities.OrderItemProductDetail, error) {
			return []entities.OrderItemProductDetail{{
				OrderId:     orderID,
				UnitId:      unitID,
				Quantity:    4,
				CreatedDate: time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC),
				Product: entities.Product{
					Name:              "Drug B",
					Unit:              "TAB",
					DrugRegistrations: []string{"KHY10"},
					DrugInfo:          &entities.DrugInfo{GenericName: "Gen B"},
				},
			}}, nil
		},
	}

	productRepo := &khy9ProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) { return nil, nil },
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return nil, errors.New("single unit lookup should not be used")
		},
		getUnitsByIDsFn: func(ids []string) ([]entities.ProductUnit, error) {
			return []entities.ProductUnit{{Id: unitID, Unit: "BOX"}}, nil
		},
	}

	items, err := getSalesReportItems(orderRepo, productRepo, branchID, req, "KHY10")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 1 || items[0].Unit != "BOX" {
		t.Fatalf("expected one item with BOX unit, got %+v", items)
	}
	if productRepo.batchUnitsCalls != 1 || productRepo.singleUnitCalls != 0 {
		t.Fatalf("expected batch product-unit lookup once and no single lookups, got batch=%d single=%d", productRepo.batchUnitsCalls, productRepo.singleUnitCalls)
	}
}
