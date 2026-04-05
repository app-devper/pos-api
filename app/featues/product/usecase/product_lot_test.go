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

type productLotRepoStub struct {
	repositories.IProduct
	getExpiringProductStocksFn func(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLotDetail, error)
	getProductLotsExpireFn     func(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error)
	getProductLotsFn           func(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLot, error)
	getProductLotByIDFn        func(id string, branchId string) (*entities.ProductLot, error)
	createProductLotFn         func(param request.ProductLot) (*entities.ProductLot, error)
	updateProductLotByIDFn     func(id string, param request.UpdateProductLot, branchId string) (*entities.ProductLot, error)
	removeProductLotByIDFn     func(id string, branchId string) (*entities.ProductLot, error)
	expiringStocksCalls        int
	globalLotsCalls            int
	lastBranchID               string
	lastParam                  request.GetProductLotsExpireRange
}

func (s *productLotRepoStub) GetExpiringProductStocks(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLotDetail, error) {
	s.expiringStocksCalls++
	s.lastBranchID = branchId
	s.lastParam = param
	return s.getExpiringProductStocksFn(param, branchId)
}

func (s *productLotRepoStub) GetProductLotsExpireNotify(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error) {
	s.globalLotsCalls++
	return s.getProductLotsExpireFn(param)
}

func (s *productLotRepoStub) GetProductLots(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLot, error) {
	s.lastBranchID = branchId
	s.lastParam = param
	return s.getProductLotsFn(param, branchId)
}

func (s *productLotRepoStub) GetProductLotById(id string, branchId string) (*entities.ProductLot, error) {
	s.lastBranchID = branchId
	return s.getProductLotByIDFn(id, branchId)
}

func (s *productLotRepoStub) CreateProductLot(param request.ProductLot) (*entities.ProductLot, error) {
	return s.createProductLotFn(param)
}

func (s *productLotRepoStub) UpdateProductLotById(id string, param request.UpdateProductLot, branchId string) (*entities.ProductLot, error) {
	s.lastBranchID = branchId
	return s.updateProductLotByIDFn(id, param, branchId)
}

func (s *productLotRepoStub) RemoveProductLotById(id string, branchId string) (*entities.ProductLot, error) {
	s.lastBranchID = branchId
	return s.removeProductLotByIDFn(id, branchId)
}

func TestGetProductLotsExpireNotifyUsesBranchScopedStocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	productRepo := &productLotRepoStub{
		getExpiringProductStocksFn: func(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLotDetail, error) {
			return []entities.ProductLotDetail{}, nil
		},
		getProductLotsExpireFn: func(param request.GetProductLotsExpireRange) ([]entities.ProductLotDetail, error) {
			t.Fatalf("global product lot query should not be used")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/products/lots/expire-notify", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", branchID)

	GetProductLotsExpireNotify(productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if productRepo.expiringStocksCalls != 1 || productRepo.globalLotsCalls != 0 {
		t.Fatalf("expected branch-scoped expiring stock query once and no global lot query, got stocks=%d global=%d", productRepo.expiringStocksCalls, productRepo.globalLotsCalls)
	}
	if productRepo.lastBranchID != branchID {
		t.Fatalf("expected branchId %s, got %s", branchID, productRepo.lastBranchID)
	}
	if !productRepo.lastParam.StartDate.Before(productRepo.lastParam.EndDate) {
		t.Fatalf("expected startDate before endDate, got start=%v end=%v", productRepo.lastParam.StartDate, productRepo.lastParam.EndDate)
	}
	duration := productRepo.lastParam.EndDate.Sub(productRepo.lastParam.StartDate)
	if duration != 24*time.Hour {
		t.Fatalf("expected 24 hour expire-notify window, got %v", duration)
	}
	if productRepo.lastParam.StartDate.Hour() != 17 {
		t.Fatalf("expected UTC timestamp aligned with Bangkok day boundary, got %v", productRepo.lastParam.StartDate)
	}
}

func TestProductLotHandlersPassBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	lotID := primitive.NewObjectID().Hex()

	t.Run("list", func(t *testing.T) {
		productRepo := &productLotRepoStub{
			getProductLotsFn: func(param request.GetProductLotsExpireRange, branchId string) ([]entities.ProductLot, error) {
				return []entities.ProductLot{}, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/products/lots", nil)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("BranchId", branchID)

		GetAllLots(productRepo)(ctx)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if productRepo.lastBranchID != branchID {
			t.Fatalf("expected branchId %s, got %s", branchID, productRepo.lastBranchID)
		}
	})

	t.Run("get by id", func(t *testing.T) {
		productRepo := &productLotRepoStub{
			getProductLotByIDFn: func(id string, branchId string) (*entities.ProductLot, error) {
				return &entities.ProductLot{}, nil
			},
		}
		req := httptest.NewRequest(http.MethodGet, "/products/lots/"+lotID, nil)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "lotId", Value: lotID}}
		ctx.Set("BranchId", branchID)

		GetLotById(productRepo)(ctx)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if productRepo.lastBranchID != branchID {
			t.Fatalf("expected branchId %s, got %s", branchID, productRepo.lastBranchID)
		}
	})

	t.Run("create", func(t *testing.T) {
		var gotParam request.ProductLot
		productRepo := &productLotRepoStub{
			createProductLotFn: func(param request.ProductLot) (*entities.ProductLot, error) {
				gotParam = param
				return &entities.ProductLot{}, nil
			},
		}
		req := httptest.NewRequest(http.MethodPost, "/products/lots", strings.NewReader(`{"productId":"`+primitive.NewObjectID().Hex()+`","quantity":2,"lotNumber":"LOT-1","expireDate":"2026-12-31T00:00:00Z","costPrice":10}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Set("BranchId", branchID)
		ctx.Set("UserId", "user-1")

		CreateLot(productRepo)(ctx)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if gotParam.BranchId != branchID {
			t.Fatalf("expected branchId %s, got %s", branchID, gotParam.BranchId)
		}
		if gotParam.UpdatedBy != "user-1" {
			t.Fatalf("expected UpdatedBy user-1, got %s", gotParam.UpdatedBy)
		}
	})

	t.Run("update", func(t *testing.T) {
		var gotParam request.UpdateProductLot
		productRepo := &productLotRepoStub{
			updateProductLotByIDFn: func(id string, param request.UpdateProductLot, branchId string) (*entities.ProductLot, error) {
				gotParam = param
				return &entities.ProductLot{}, nil
			},
		}
		req := httptest.NewRequest(http.MethodPut, "/products/lots/"+lotID, strings.NewReader(`{"quantity":2,"lotNumber":"LOT-1","expireDate":"2026-12-31T00:00:00Z","costPrice":10}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "lotId", Value: lotID}}
		ctx.Set("BranchId", branchID)
		ctx.Set("UserId", "user-2")

		UpdateLotById(productRepo)(ctx)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if productRepo.lastBranchID != branchID {
			t.Fatalf("expected branchId %s, got %s", branchID, productRepo.lastBranchID)
		}
		if gotParam.BranchId != branchID {
			t.Fatalf("expected branchId %s, got %s", branchID, gotParam.BranchId)
		}
		if gotParam.UpdatedBy != "user-2" {
			t.Fatalf("expected UpdatedBy user-2, got %s", gotParam.UpdatedBy)
		}
	})

	t.Run("delete", func(t *testing.T) {
		productRepo := &productLotRepoStub{
			removeProductLotByIDFn: func(id string, branchId string) (*entities.ProductLot, error) {
				return &entities.ProductLot{}, nil
			},
		}
		req := httptest.NewRequest(http.MethodDelete, "/products/lots/"+lotID, nil)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req
		ctx.Params = gin.Params{{Key: "lotId", Value: lotID}}
		ctx.Set("BranchId", branchID)

		DeleteLotById(productRepo)(ctx)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if productRepo.lastBranchID != branchID {
			t.Fatalf("expected branchId %s, got %s", branchID, productRepo.lastBranchID)
		}
	})
}
