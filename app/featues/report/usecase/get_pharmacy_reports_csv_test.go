package usecase

import (
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

type khy9CSVReceiveStub struct {
	repositories.IReceive
	getReceivesFn          func(form request.GetReceiveRange) ([]entities.Receive, error)
	getReceiveItemsByIDFn  func(receiveId string) ([]entities.ReceiveItem, error)
	getReceiveItemsByIDsFn func(receiveIds []string) ([]entities.ReceiveItem, error)
	singleItemCalls        int
	batchItemCalls         int
}

func (s *khy9CSVReceiveStub) GetReceives(form request.GetReceiveRange) ([]entities.Receive, error) {
	return s.getReceivesFn(form)
}

func (s *khy9CSVReceiveStub) GetReceiveItemsByReceiveId(receiveId string) ([]entities.ReceiveItem, error) {
	s.singleItemCalls++
	return s.getReceiveItemsByIDFn(receiveId)
}

func (s *khy9CSVReceiveStub) GetReceiveItemsByReceiveIds(receiveIds []string) ([]entities.ReceiveItem, error) {
	s.batchItemCalls++
	return s.getReceiveItemsByIDsFn(receiveIds)
}

func TestGetKHY9CSVUsesReceiveItemUnit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	unitID := primitive.NewObjectID().Hex()
	supplierID := primitive.NewObjectID()

	receiveRepo := &khy9CSVReceiveStub{
		getReceivesFn: func(form request.GetReceiveRange) ([]entities.Receive, error) {
			return []entities.Receive{
				{Id: receiveID, Code: "RC-001", SupplierId: supplierID, CreatedDate: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)},
			}, nil
		},
		getReceiveItemsByIDFn: func(receiveId string) ([]entities.ReceiveItem, error) {
			t.Fatalf("single receive-item lookup should not be used")
			return nil, nil
		},
		getReceiveItemsByIDsFn: func(receiveIds []string) ([]entities.ReceiveItem, error) {
			return []entities.ReceiveItem{
				{ReceiveId: receiveID, ProductId: productID, UnitId: unitID, Quantity: 2, CostPrice: 10, LotNumber: "LOT-1"},
			}, nil
		},
	}

	productRepo := &khy9ProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return []entities.Product{
				{Id: productID, Name: "Drug A", Unit: "TAB", DrugRegistrations: []string{"KHY9"}, DrugInfo: &entities.DrugInfo{GenericName: "Gen A"}},
			}, nil
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			t.Fatalf("single product-unit lookup should not be used")
			return nil, nil
		},
		getUnitsByIDsFn: func(ids []string) ([]entities.ProductUnit, error) {
			unitID, _ := primitive.ObjectIDFromHex(ids[0])
			return []entities.ProductUnit{{Id: unitID, Unit: "BOX"}}, nil
		},
	}

	supplierRepo := &khy9SupplierStub{
		getSupplierByIDFn: func(id string) (*entities.Supplier, error) {
			t.Fatalf("single supplier lookup should not be used")
			return nil, nil
		},
		getSuppliersByIDsFn: func(ids []string) ([]entities.Supplier, error) {
			return []entities.Supplier{{Id: supplierID, Name: "Supplier A"}}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/pharmacy/khy9/csv?startDate=2026-01-01T00:00:00Z&endDate=2026-02-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetKHY9CSV(receiveRepo, productRepo, supplierRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !strings.Contains(w.Body.String(), "BOX") {
		t.Fatalf("expected CSV to contain resolved unit BOX, got %s", w.Body.String())
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
