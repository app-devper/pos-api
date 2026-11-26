package usecase

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const khy9RangeQuery = "?startDate=2026-01-01T00:00:00Z&endDate=2026-02-01T00:00:00Z"

// khy9ReposWithACancelledReceipt returns repositories holding one Receive per
// status (code = status label), each with one KHY9 line of quantity 1.
func khy9ReposWithACancelledReceipt(t *testing.T) (*khy9ReceiveStub, *khy9ProductStub, *khy9SupplierStub) {
	t.Helper()
	productID := primitive.NewObjectID()
	supplierID := primitive.NewObjectID()
	statuses := map[string]string{"imported": constant.IMPORTED, "cancelled": constant.CANCELLED, "draft": constant.ACTIVE, "legacy": ""}
	receives := make([]entities.Receive, 0, len(statuses))
	for code, status := range statuses {
		receives = append(receives, entities.Receive{Id: primitive.NewObjectID(), Code: code, Status: status, SupplierId: supplierID, CreatedDate: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)})
	}
	receiveRepo := &khy9ReceiveStub{
		getReceivesFn: func(form request.GetReceiveRange) ([]entities.Receive, error) { return receives, nil },
		getReceiveItemsByIDsFn: func(ids []string) ([]entities.ReceiveItem, error) {
			items := make([]entities.ReceiveItem, 0, len(ids))
			for _, id := range ids {
				receiveID, _ := primitive.ObjectIDFromHex(id)
				items = append(items, entities.ReceiveItem{ReceiveId: receiveID, ProductId: productID, Quantity: 1, CostPrice: 10})
			}
			return items, nil
		},
	}
	productRepo := &khy9ProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return []entities.Product{{Id: productID, Name: "Drug A", Unit: "TAB", DrugRegistrations: []string{"KHY9"}}}, nil
		},
		getUnitsByIDsFn: func(ids []string) ([]entities.ProductUnit, error) { return nil, nil },
	}
	supplierRepo := &khy9SupplierStub{
		getSuppliersByIDsFn: func(ids []string) ([]entities.Supplier, error) {
			return []entities.Supplier{{Id: supplierID, Name: "Supplier A"}}, nil
		},
	}
	return receiveRepo, productRepo, supplierRepo
}

func serveKHY9(handler gin.HandlerFunc, url string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, url, nil)
	ctx.Set("BranchId", primitive.NewObjectID().Hex())
	handler(ctx)
	return w
}

func TestGetKHY9DataLeavesOutACancelledReceipt(t *testing.T) {
	receiveRepo, productRepo, supplierRepo := khy9ReposWithACancelledReceipt(t)

	w := serveKHY9(GetKHY9Data(receiveRepo, productRepo, supplierRepo), "/reports/pharmacy/khy9/data"+khy9RangeQuery)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp pharmacyReportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, item := range resp.Items {
		codes[item.Code] = true
	}
	if len(resp.Items) != 3 || codes["cancelled"] || !codes["imported"] || !codes["draft"] || !codes["legacy"] {
		t.Fatalf("items %+v", resp.Items)
	}
}

func TestGetKHY9CSVLeavesOutACancelledReceipt(t *testing.T) {
	receiveRepo, productRepo, supplierRepo := khy9ReposWithACancelledReceipt(t)

	w := serveKHY9(GetKHY9CSV(receiveRepo, productRepo, supplierRepo), "/reports/pharmacy/khy9/csv"+khy9RangeQuery)

	body := w.Body.String()
	if w.Code != http.StatusOK || strings.Contains(body, "cancelled") {
		t.Fatalf("status %d body %s", w.Code, body)
	}
	for _, code := range []string{"imported", "draft", "legacy"} {
		if !strings.Contains(body, code) {
			t.Fatalf("CSV lacks receipt %s: %s", code, body)
		}
	}
}
