package usecase

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type expiringProductStub struct {
	repositories.IProduct
	getExpiringProductStocksFn func(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLotDetail, error)
	getProductLotsExpireFn     func(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error)
	expiringStocksCalls        int
	globalLotsCalls            int
	lastBranchID               string
	lastParam                  request.GetProductLotsExpireRange
}

func (s *expiringProductStub) GetExpiringProductStocks(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLotDetail, error) {
	s.expiringStocksCalls++
	s.lastBranchID = branchId
	s.lastParam = param
	return s.getExpiringProductStocksFn(param, branchId)
}

func (s *expiringProductStub) GetProductLotsExpireNotify(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error) {
	s.globalLotsCalls++
	return s.getProductLotsExpireFn(param)
}

func TestGetExpiringProductsUsesBranchScopedStocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	productRepo := &expiringProductStub{
		getExpiringProductStocksFn: func(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLotDetail, error) {
			return []entities.ProductLotDetail{}, nil
		},
		getProductLotsExpireFn: func(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error) {
			t.Fatalf("global product lot query should not be used")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard/expiring", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", branchID)

	before := time.Now()
	GetExpiringProducts(productRepo)(ctx)
	after := time.Now()

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if productRepo.expiringStocksCalls != 1 || productRepo.globalLotsCalls != 0 {
		t.Fatalf("expected branch-scoped expiring stock query once and no global lot query, got stocks=%d global=%d", productRepo.expiringStocksCalls, productRepo.globalLotsCalls)
	}
	if productRepo.lastBranchID != branchID {
		t.Fatalf("expected branchId %s, got %s", branchID, productRepo.lastBranchID)
	}
	if productRepo.lastParam.StartDate.Before(before.Add(-time.Second)) || productRepo.lastParam.StartDate.After(after.Add(time.Second)) {
		t.Fatalf("unexpected startDate: %+v", productRepo.lastParam.StartDate)
	}
	expectedEndLower := productRepo.lastParam.StartDate.AddDate(0, 6, 0).Add(-time.Second)
	expectedEndUpper := productRepo.lastParam.StartDate.AddDate(0, 6, 0).Add(time.Second)
	if productRepo.lastParam.EndDate.Before(expectedEndLower) || productRepo.lastParam.EndDate.After(expectedEndUpper) {
		t.Fatalf("expected endDate about six months after startDate, got start=%v end=%v", productRepo.lastParam.StartDate, productRepo.lastParam.EndDate)
	}
}
