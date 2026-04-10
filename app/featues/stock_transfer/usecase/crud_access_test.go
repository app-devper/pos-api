package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type transferAccessRepoStub struct {
	repositories.IStockTransfer
	getByIDFn func(id string) (*entities.StockTransfer, error)
	approveFn func(id string, updatedBy string) (*entities.StockTransfer, error)
	rejectFn  func(id string, updatedBy string) (*entities.StockTransfer, error)
	createFn  func(form request.StockTransfer) (*entities.StockTransfer, error)
}

func (s *transferAccessRepoStub) GetStockTransferById(id string) (*entities.StockTransfer, error) {
	return s.getByIDFn(id)
}

func (s *transferAccessRepoStub) ApproveStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error) {
	return s.approveFn(id, updatedBy)
}

func TestGetStockTransferByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &transferAccessRepoStub{
		getByIDFn: func(id string) (*entities.StockTransfer, error) {
			return &entities.StockTransfer{
				Id:           primitive.NewObjectID(),
				FromBranchId: primitive.NewObjectID(),
				ToBranchId:   primitive.NewObjectID(),
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/stock-transfers/1", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetStockTransferById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.SY_FORBIDDEN_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.SY_FORBIDDEN_002, w.Body.String())
	}
}

func TestApproveStockTransferRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &transferAccessRepoStub{
		getByIDFn: func(id string) (*entities.StockTransfer, error) {
			return &entities.StockTransfer{
				Id:           primitive.NewObjectID(),
				FromBranchId: primitive.NewObjectID(),
				ToBranchId:   primitive.NewObjectID(),
				Status:       "PENDING",
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/stock-transfers/1/approve", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	ApproveStockTransfer(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.SY_FORBIDDEN_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.SY_FORBIDDEN_002, w.Body.String())
	}
}
