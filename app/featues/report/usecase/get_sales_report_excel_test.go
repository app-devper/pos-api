package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type salesExcelOrderStub struct {
	repositories.IOrder
	getOrderRangeFn func(form request.GetOrderRange) ([]entities.Order, error)
}

func (s *salesExcelOrderStub) GetOrderRange(form request.GetOrderRange) ([]entities.Order, error) {
	return s.getOrderRangeFn(form)
}

func TestGetSalesReportExcelSkipsInactiveOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orderRepo := &salesExcelOrderStub{
		getOrderRangeFn: func(form request.GetOrderRange) ([]entities.Order, error) {
			return []entities.Order{
				{
					Id:           primitive.NewObjectID(),
					Code:         "OR-001",
					CreatedDate:  time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC),
					CustomerCode: "C001",
					CustomerName: "Active Customer",
					Type:         "cash",
					Status:       constant.ACTIVE,
					Total:        100,
					TotalCost:    60,
				},
				{
					Id:           primitive.NewObjectID(),
					Code:         "OR-002",
					CreatedDate:  time.Date(2026, 1, 2, 11, 0, 0, 0, time.UTC),
					CustomerCode: "C002",
					CustomerName: "Cancelled Customer",
					Type:         "cash",
					Status:       "CANCELLED",
					Total:        200,
					TotalCost:    50,
				},
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/sales/excel?startDate=2026-01-01T00:00:00Z&endDate=2026-02-01T00:00:00Z", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetSalesReportExcel(orderRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Content-Type"); !strings.Contains(got, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") {
		t.Fatalf("expected excel content type, got %s", got)
	}
}
