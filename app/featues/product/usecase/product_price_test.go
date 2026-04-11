package usecase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productPriceRepoStub struct {
	repositories.IProduct
	createPriceCascadeFn func(param request.ProductPrice, branchId string, userId string) (*entities.ProductPrice, error)
	removePriceCascadeFn func(id string, branchId string, userId string) (*entities.ProductPrice, error)
	updatePriceByIDFn    func(id string, param request.ProductPrice) (*entities.ProductPrice, error)
	getUnitByIDFn        func(id string) (*entities.ProductUnit, error)
	createHistoryFn      func(param request.ProductHistory) (*entities.ProductHistory, error)
}

func (s *productPriceRepoStub) CreateProductPriceCascade(param request.ProductPrice, branchId string, userId string) (*entities.ProductPrice, error) {
	return s.createPriceCascadeFn(param, branchId, userId)
}

func (s *productPriceRepoStub) RemoveProductPriceCascade(id string, branchId string, userId string) (*entities.ProductPrice, error) {
	return s.removePriceCascadeFn(id, branchId, userId)
}

func (s *productPriceRepoStub) UpdateProductPriceById(id string, param request.ProductPrice) (*entities.ProductPrice, error) {
	return s.updatePriceByIDFn(id, param)
}

func (s *productPriceRepoStub) GetProductUnitById(id string) (*entities.ProductUnit, error) {
	return s.getUnitByIDFn(id)
}

func (s *productPriceRepoStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

type productPriceStockStub struct {
	repositories.IProductStock
	createHistoryFn func(param request.ProductHistory) (*entities.ProductHistory, error)
}

func (s *productPriceStockStub) CreateProductHistory(param request.ProductHistory) (*entities.ProductHistory, error) {
	return s.createHistoryFn(param)
}

func TestCreateProductPriceUsesTransactionalCascadeMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	unitID := primitive.NewObjectID().Hex()
	var gotReq request.ProductPrice
	var gotBranchID string
	var gotUserID string

	repo := &productPriceRepoStub{
		createPriceCascadeFn: func(param request.ProductPrice, branchId string, userId string) (*entities.ProductPrice, error) {
			gotReq = param
			gotBranchID = branchId
			gotUserID = userId
			return &entities.ProductPrice{Id: primitive.NewObjectID()}, nil
		},
	}

	body := `{"productId":"` + productID + `","unitId":"` + unitID + `","price":45,"customerType":"General"}`
	req := httptest.NewRequest(http.MethodPost, "/product-prices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateProductPrice(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotReq.ProductId != productID || gotReq.UnitId != unitID || gotReq.Price != 45 {
		t.Fatalf("expected request payload to be forwarded, got %+v", gotReq)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
	if gotUserID != "user-1" {
		t.Fatalf("expected user id user-1, got %s", gotUserID)
	}
}

func TestRemoveProductPriceByIdUsesTransactionalCascadeMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	priceID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	var gotID string
	var gotBranchID string
	var gotUserID string

	repo := &productPriceRepoStub{
		removePriceCascadeFn: func(id string, branchId string, userId string) (*entities.ProductPrice, error) {
			gotID = id
			gotBranchID = branchId
			gotUserID = userId
			return &entities.ProductPrice{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/product-prices/"+priceID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "priceId", Value: priceID}}
	ctx.Set("UserId", "user-2")
	ctx.Set("BranchId", branchID)

	RemoveProductPriceById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != priceID {
		t.Fatalf("expected price id %s, got %s", priceID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
	if gotUserID != "user-2" {
		t.Fatalf("expected user id user-2, got %s", gotUserID)
	}
}

func TestUpdateProductPriceByIdSkipsHistoryWhenUnitLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	priceID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	unitID := primitive.NewObjectID().Hex()
	historyCalled := false

	repo := &productPriceRepoStub{
		updatePriceByIDFn: func(id string, param request.ProductPrice) (*entities.ProductPrice, error) {
			return &entities.ProductPrice{Id: primitive.NewObjectID()}, nil
		},
		getUnitByIDFn: func(id string) (*entities.ProductUnit, error) {
			return nil, errors.New("unit lookup failed")
		},
		createHistoryFn: func(param request.ProductHistory) (*entities.ProductHistory, error) {
			historyCalled = true
			return &entities.ProductHistory{}, nil
		},
	}

	body := `{"productId":"` + productID + `","unitId":"` + unitID + `","price":45,"customerType":"General"}`
	req := httptest.NewRequest(http.MethodPatch, "/product-prices/"+priceID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "priceId", Value: priceID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	stockRepo := &productPriceStockStub{
		createHistoryFn: repo.createHistoryFn,
	}

	UpdateProductPriceById(repo, stockRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if historyCalled {
		t.Fatal("expected history creation to be skipped when unit lookup fails")
	}
}
