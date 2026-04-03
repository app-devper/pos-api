package usecase

import (
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
}

func (s *productPriceRepoStub) CreateProductPriceCascade(param request.ProductPrice, branchId string, userId string) (*entities.ProductPrice, error) {
	return s.createPriceCascadeFn(param, branchId, userId)
}

func (s *productPriceRepoStub) RemoveProductPriceCascade(id string, branchId string, userId string) (*entities.ProductPrice, error) {
	return s.removePriceCascadeFn(id, branchId, userId)
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
