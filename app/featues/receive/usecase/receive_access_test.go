package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type receiveAccessRepoStub struct {
	repositories.IReceive
	getReceiveByIDFn          func(id string) (*entities.Receive, error)
	getReceiveItemsByIDFn     func(receiveId string) ([]entities.ReceiveItem, error)
	updateReceiveStatusByIDFn func(id string, status string, updatedBy string) (*entities.Receive, error)
	importReceiveToStockFn    func(receiveId string, userId string, branchId string) (*entities.Receive, error)
	updateReceiveItemsByIDFn  func(id string, form request.UpdateReceiveItems) (*entities.Receive, error)
	updateReceiveTotalCostFn  func(id string, totalCost float64) (*entities.Receive, error)
}

func (s *receiveAccessRepoStub) GetReceiveById(id string) (*entities.Receive, error) {
	return s.getReceiveByIDFn(id)
}

func (s *receiveAccessRepoStub) GetReceiveItemsByReceiveId(receiveId string) ([]entities.ReceiveItem, error) {
	if s.getReceiveItemsByIDFn != nil {
		return s.getReceiveItemsByIDFn(receiveId)
	}
	return nil, nil
}

func (s *receiveAccessRepoStub) UpdateReceiveStatusById(id string, status string, updatedBy string) (*entities.Receive, error) {
	return s.updateReceiveStatusByIDFn(id, status, updatedBy)
}

func (s *receiveAccessRepoStub) ImportReceiveToStock(receiveId string, userId string, branchId string) (*entities.Receive, error) {
	return s.importReceiveToStockFn(receiveId, userId, branchId)
}

func (s *receiveAccessRepoStub) UpdateReceiveItemsById(id string, form request.UpdateReceiveItems) (*entities.Receive, error) {
	return s.updateReceiveItemsByIDFn(id, form)
}

func (s *receiveAccessRepoStub) UpdateReceiveTotalCostById(id string, totalCost float64) (*entities.Receive, error) {
	return s.updateReceiveTotalCostFn(id, totalCost)
}

func TestGetReceiveByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &receiveAccessRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/receives/1", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetReceiveById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.SY_FORBIDDEN_002) {
		t.Fatalf("expected errcode %s, got %s", errcode.SY_FORBIDDEN_002, w.Body.String())
	}
}

func TestDeleteReceiveByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &receiveAccessRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		updateReceiveStatusByIDFn: func(id string, status string, updatedBy string) (*entities.Receive, error) {
			t.Fatal("delete should not be called for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/receives/1", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	DeleteReceiveById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestImportReceiveToStockRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &receiveAccessRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		importReceiveToStockFn: func(receiveId string, userId string, branchId string) (*entities.Receive, error) {
			t.Fatal("import should not be called for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/receives/1/import", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	ImportReceiveToStock(repo, nil)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestUpdateReceiveItemsByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &receiveAccessRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		updateReceiveItemsByIDFn: func(id string, form request.UpdateReceiveItems) (*entities.Receive, error) {
			t.Fatal("update items should not be called for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/receives/1/items", strings.NewReader(`{"items":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	UpdateReceiveItemsById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestUpdateReceiveTotalCostByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &receiveAccessRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID()}, nil
		},
		updateReceiveTotalCostFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/receives/1/total-cost", strings.NewReader(`{"totalCost":10}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	UpdateReceiveTotalCostById(repo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestUpdateReceiveItemsByIdRejectsImportedReceive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID()
	repo := &receiveAccessRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Status: constant.IMPORTED}, nil
		},
		updateReceiveItemsByIDFn: func(id string, form request.UpdateReceiveItems) (*entities.Receive, error) {
			t.Fatal("update items should not be called for imported receive")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/receives/1/items", strings.NewReader(`{"items":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	UpdateReceiveItemsById(repo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "cannot modify imported receive") {
		t.Fatalf("expected imported receive guard, got %s", w.Body.String())
	}
}

func TestUpdateReceiveTotalCostByIdRejectsImportedReceive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID()
	repo := &receiveAccessRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Status: constant.IMPORTED}, nil
		},
		updateReceiveTotalCostFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called for imported receive")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPatch, "/receives/1/total-cost", strings.NewReader(`{"totalCost":10}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("BranchId", branchID.Hex())

	UpdateReceiveTotalCostById(repo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "cannot modify imported receive") {
		t.Fatalf("expected imported receive guard, got %s", w.Body.String())
	}
}
