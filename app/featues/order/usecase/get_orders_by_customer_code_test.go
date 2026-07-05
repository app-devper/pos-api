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

type orderCustomerCodeRepoStub struct {
	repositories.IOrder
	getOrdersByCustomerCodeFn func(customerCode string, branchId string) ([]entities.Order, error)
}

func (s *orderCustomerCodeRepoStub) GetOrdersByCustomerCode(customerCode string, branchId string) ([]entities.Order, error) {
	return s.getOrdersByCustomerCodeFn(customerCode, branchId)
}

func TestGetOrdersByCustomerCodePassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customerCode := "C001"
	branchID := primitive.NewObjectID().Hex()
	var gotCustomerCode string
	var gotBranchID string

	repo := &orderCustomerCodeRepoStub{
		getOrdersByCustomerCodeFn: func(code string, branchId string) ([]entities.Order, error) {
			gotCustomerCode = code
			gotBranchID = branchId
			return []entities.Order{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/orders/customers/"+customerCode, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "customerCode", Value: customerCode}}
	ctx.Set("BranchId", branchID)

	GetOrdersByCustomerCode(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotCustomerCode != customerCode {
		t.Fatalf("expected customer code %s, got %s", customerCode, gotCustomerCode)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
}
