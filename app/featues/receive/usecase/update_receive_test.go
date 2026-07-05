package usecase

import (
	"errors"
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

type updateReceiveRepoStub struct {
	repositories.IReceive
	getReceiveByIDFn func(id string) (*entities.Receive, error)
	updateByIDFn     func(id string, form request.UpdateReceive) (*entities.Receive, error)
}

func (s *updateReceiveRepoStub) GetReceiveById(id string) (*entities.Receive, error) {
	return s.getReceiveByIDFn(id)
}

func (s *updateReceiveRepoStub) UpdateReceiveById(id string, form request.UpdateReceive) (*entities.Receive, error) {
	return s.updateByIDFn(id, form)
}

type updateReceiveProductStub struct {
	repositories.IProduct
	getProductByIDFn func(id string) (*entities.Product, error)
}

func (s *updateReceiveProductStub) GetProductById(id string) (*entities.Product, error) {
	return s.getProductByIDFn(id)
}

func TestUpdateReceiveByIdFiltersInvalidItemsBeforeTransactionalUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	validProductID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()
	var gotForm request.UpdateReceive

	receiveRepo := &updateReceiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Code: "RC-001"}, nil
		},
		updateByIDFn: func(id string, form request.UpdateReceive) (*entities.Receive, error) {
			gotForm = form
			return &entities.Receive{Items: []entities.ReceiveItem{{}}}, nil
		},
	}
	productRepo := &updateReceiveProductStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return &entities.Product{
				Id:           primitive.NewObjectID(),
				Name:         "Paracetamol",
				SerialNumber: "PD-001",
				Price:        10,
				Unit:         "TAB",
			}, nil
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","reference":"ref-1","items":[{"productId":"","costPrice":1,"quantity":1},{"productId":"` + validProductID + `","costPrice":5,"quantity":2,"expireDate":"2026-12-31"}]}`
	req := httptest.NewRequest(http.MethodPut, "/receives/"+receiveID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	UpdateReceiveById(receiveRepo, productRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if len(gotForm.ReceiveItems) != 1 {
		t.Fatalf("expected only 1 valid item sent to repository, got %d", len(gotForm.ReceiveItems))
	}
	if gotForm.TotalCost != 10 {
		t.Fatalf("expected total cost 10, got %f", gotForm.TotalCost)
	}
	if gotForm.UpdatedBy != "user-1" {
		t.Fatalf("expected UpdatedBy user-1, got %s", gotForm.UpdatedBy)
	}
}

func TestUpdateReceiveByIdFailsWhenProductLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()

	receiveRepo := &updateReceiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Code: "RC-001"}, nil
		},
		updateByIDFn: func(id string, form request.UpdateReceive) (*entities.Receive, error) {
			t.Fatal("update receive should not be called when product lookup fails")
			return nil, nil
		},
	}
	productRepo := &updateReceiveProductStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, errors.New("product lookup failed")
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","reference":"ref-1","items":[{"productId":"` + productID + `","costPrice":5,"quantity":2}]}`
	req := httptest.NewRequest(http.MethodPut, "/receives/"+receiveID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	UpdateReceiveById(receiveRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.RC_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s, got %s", errcode.RC_BAD_REQUEST_002, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "product lookup failed") {
		t.Fatalf("expected product lookup error, got %s", w.Body.String())
	}
}

func TestUpdateReceiveByIdFailsWhenProductNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID().Hex()
	productID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID()

	receiveRepo := &updateReceiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Code: "RC-001"}, nil
		},
		updateByIDFn: func(id string, form request.UpdateReceive) (*entities.Receive, error) {
			t.Fatal("update receive should not be called when product is missing")
			return nil, nil
		},
	}
	productRepo := &updateReceiveProductStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, nil
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","reference":"ref-1","items":[{"productId":"` + productID + `","costPrice":5,"quantity":2}]}`
	req := httptest.NewRequest(http.MethodPut, "/receives/"+receiveID, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	UpdateReceiveById(receiveRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "product "+productID+" not found") {
		t.Fatalf("expected product not found error, got %s", w.Body.String())
	}
}

func TestUpdateReceiveByIdRejectsForeignBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveRepo := &updateReceiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID(), Code: "RC-001"}, nil
		},
		updateByIDFn: func(id string, form request.UpdateReceive) (*entities.Receive, error) {
			t.Fatal("update receive should not be called for foreign branch")
			return nil, nil
		},
	}
	productRepo := &updateReceiveProductStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			t.Fatal("product lookup should not be called for foreign branch")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/receives/"+primitive.NewObjectID().Hex(), strings.NewReader(`{"supplierId":"`+primitive.NewObjectID().Hex()+`","items":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	UpdateReceiveById(receiveRepo, productRepo)(ctx)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.SY_FORBIDDEN_002) {
		t.Fatalf("expected errcode %s, got %s", errcode.SY_FORBIDDEN_002, w.Body.String())
	}
}

func TestUpdateReceiveByIdFailsWhenExpireDateIsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveRepo := &updateReceiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: primitive.NewObjectID(), Code: "RC-001"}, nil
		},
		updateByIDFn: func(id string, form request.UpdateReceive) (*entities.Receive, error) {
			t.Fatal("update receive should not be called when expire date is invalid")
			return nil, nil
		},
	}
	productRepo := &updateReceiveProductStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return &entities.Product{
				Id:           primitive.NewObjectID(),
				Name:         "Paracetamol",
				SerialNumber: "PD-001",
				Price:        10,
				Unit:         "TAB",
			}, nil
		},
	}

	branchID := primitive.NewObjectID()
	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","reference":"ref-1","items":[{"productId":"` + primitive.NewObjectID().Hex() + `","costPrice":5,"quantity":2,"expireDate":"31/12/2026"}]}`
	req := httptest.NewRequest(http.MethodPut, "/receives/"+primitive.NewObjectID().Hex(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	receiveRepo.getReceiveByIDFn = func(id string) (*entities.Receive, error) {
		return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Code: "RC-001"}, nil
	}

	UpdateReceiveById(receiveRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid expireDate") {
		t.Fatalf("expected invalid expireDate error, got %s", w.Body.String())
	}
}

func TestUpdateReceiveByIdRejectsImportedReceive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID()
	receiveRepo := &updateReceiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), BranchId: branchID, Code: "RC-001", Status: constant.IMPORTED}, nil
		},
		updateByIDFn: func(id string, form request.UpdateReceive) (*entities.Receive, error) {
			t.Fatal("update receive should not be called for imported receive")
			return nil, nil
		},
	}
	productRepo := &updateReceiveProductStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			t.Fatal("product lookup should not be called for imported receive")
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/receives/"+primitive.NewObjectID().Hex(), strings.NewReader(`{"supplierId":"`+primitive.NewObjectID().Hex()+`","items":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: primitive.NewObjectID().Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	UpdateReceiveById(receiveRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "cannot modify imported receive") {
		t.Fatalf("expected imported receive guard, got %s", w.Body.String())
	}
}
