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

type productUnitRepoStub struct {
	repositories.IProduct
	createUnitCascadeFn func(param request.CreateProductUnit, branchId string, userId string) (*entities.ProductUnit, error)
	removeUnitCascadeFn func(id string, branchId string, userId string) (*entities.ProductUnit, error)
}

func (s *productUnitRepoStub) CreateProductUnitCascade(param request.CreateProductUnit, branchId string, userId string) (*entities.ProductUnit, error) {
	return s.createUnitCascadeFn(param, branchId, userId)
}

func (s *productUnitRepoStub) RemoveProductUnitCascade(id string, branchId string, userId string) (*entities.ProductUnit, error) {
	return s.removeUnitCascadeFn(id, branchId, userId)
}

func TestCreateProductUnitUsesTransactionalCascadeMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	var gotReq request.CreateProductUnit
	var gotBranchID string
	var gotUserID string

	repo := &productUnitRepoStub{
		createUnitCascadeFn: func(param request.CreateProductUnit, branchId string, userId string) (*entities.ProductUnit, error) {
			gotReq = param
			gotBranchID = branchId
			gotUserID = userId
			return &entities.ProductUnit{Id: primitive.NewObjectID(), Unit: param.Unit}, nil
		},
	}

	body := `{"productId":"` + productID + `","unit":"BOX","size":12,"costPrice":24,"price":30,"volume":1.5,"volumeUnit":"L","barcode":"BOX-001"}`
	req := httptest.NewRequest(http.MethodPost, "/product-units", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateProductUnit(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotReq.ProductId != productID || gotReq.Unit != "BOX" || gotReq.Price != 30 {
		t.Fatalf("expected request payload to be forwarded, got %+v", gotReq)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
	if gotUserID != "user-1" {
		t.Fatalf("expected user id user-1, got %s", gotUserID)
	}
}

func TestRemoveProductUnitByIdUsesTransactionalCascadeMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	unitID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID()

	var gotID string
	var gotBranchID string
	var gotUserID string

	repo := &productUnitRepoStub{
		removeUnitCascadeFn: func(id string, branchId string, userId string) (*entities.ProductUnit, error) {
			gotID = id
			gotBranchID = branchId
			gotUserID = userId
			return &entities.ProductUnit{Id: primitive.NewObjectID(), ProductId: productID, Unit: "BOX"}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/product-units/"+unitID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "unitId", Value: unitID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	RemoveProductUnitById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != unitID {
		t.Fatalf("expected unit id %s, got %s", unitID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
	if gotUserID != "user-1" {
		t.Fatalf("expected user id user-1, got %s", gotUserID)
	}
}
