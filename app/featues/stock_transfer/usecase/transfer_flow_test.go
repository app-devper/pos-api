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
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type stockTransferRepoStub struct {
	repositories.IStockTransfer
	createFn       func(form request.StockTransfer) (*entities.StockTransfer, error)
	getByIDFn      func(id string) (*entities.StockTransfer, error)
	updateStatusFn func(id string, form request.UpdateStockTransfer) (*entities.StockTransfer, error)
	approveFn      func(id string, updatedBy string) (*entities.StockTransfer, error)
	rejectFn       func(id string, updatedBy string) (*entities.StockTransfer, error)
}

func (s *stockTransferRepoStub) CreateStockTransfer(form request.StockTransfer) (*entities.StockTransfer, error) {
	return s.createFn(form)
}

func (s *stockTransferRepoStub) CreateStockTransferWithReservation(form request.StockTransfer) (*entities.StockTransfer, error) {
	return s.createFn(form)
}

func (s *stockTransferRepoStub) GetStockTransferById(id string) (*entities.StockTransfer, error) {
	return s.getByIDFn(id)
}

func (s *stockTransferRepoStub) UpdateStockTransferStatus(id string, form request.UpdateStockTransfer) (*entities.StockTransfer, error) {
	return s.updateStatusFn(id, form)
}

func (s *stockTransferRepoStub) ApproveStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error) {
	return s.approveFn(id, updatedBy)
}

func (s *stockTransferRepoStub) RejectStockTransfer(id string, updatedBy string) (*entities.StockTransfer, error) {
	return s.rejectFn(id, updatedBy)
}

type transferProductStub struct {
	repositories.IProduct
	removeStockFn   func(stockId string, quantity int) (*entities.ProductStock, error)
	addStockFn      func(stockId string, quantity int) (*entities.ProductStock, error)
	getStockByIDFn  func(id string) (*entities.ProductStock, error)
	createStockFn   func(param request.ProductStock) (*entities.ProductStock, error)
	removeStockByID func(id string) (*entities.ProductStock, error)
}

func (s *transferProductStub) RemoveProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.removeStockFn(stockId, quantity)
}

func (s *transferProductStub) AddProductStockQuantityById(stockId string, quantity int) (*entities.ProductStock, error) {
	return s.addStockFn(stockId, quantity)
}

func (s *transferProductStub) GetProductStockById(id string) (*entities.ProductStock, error) {
	return s.getStockByIDFn(id)
}

func (s *transferProductStub) CreateProductStock(param request.ProductStock) (*entities.ProductStock, error) {
	return s.createStockFn(param)
}

func (s *transferProductStub) RemoveProductStockById(id string) (*entities.ProductStock, error) {
	return s.removeStockByID(id)
}

type transferSequenceStub struct {
	repositories.ISequence
	nextFn func(field string) (*entities.Sequence, error)
}

func (s *transferSequenceStub) NextSequence(field string) (*entities.Sequence, error) {
	if s.nextFn != nil {
		return s.nextFn(field)
	}
	return &entities.Sequence{Field: field, Prefix: "TR", Value: 1, Format: 4}, nil
}

func TestCreateStockTransferReturnsErrorWhenTransactionalCreateFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sourceBranchID := primitive.NewObjectID()
	targetBranchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	stockID := primitive.NewObjectID()

	repo := &stockTransferRepoStub{
		createFn: func(form request.StockTransfer) (*entities.StockTransfer, error) {
			return nil, errors.New("insert failed")
		},
	}
	productRepo := &transferProductStub{}

	body := `{"toBranchId":"` + targetBranchID.Hex() + `","items":[{"productId":"` + productID.Hex() + `","stockId":"` + stockID.Hex() + `","quantity":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/stock-transfers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", sourceBranchID.Hex())

	CreateStockTransfer(repo, productRepo, &transferSequenceStub{})(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.TR_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.TR_BAD_REQUEST_002, w.Body.String())
	}
}

func TestCreateStockTransferFailsWhenSequenceLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sourceBranchID := primitive.NewObjectID()
	targetBranchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	stockID := primitive.NewObjectID()

	repo := &stockTransferRepoStub{
		createFn: func(form request.StockTransfer) (*entities.StockTransfer, error) {
			t.Fatal("create stock transfer should not be called when sequence fails")
			return nil, nil
		},
	}
	productRepo := &transferProductStub{}
	sequenceRepo := &transferSequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return nil, errors.New("sequence failed")
		},
	}

	body := `{"toBranchId":"` + targetBranchID.Hex() + `","items":[{"productId":"` + productID.Hex() + `","stockId":"` + stockID.Hex() + `","quantity":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/stock-transfers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", sourceBranchID.Hex())

	CreateStockTransfer(repo, productRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "sequence failed") {
		t.Fatalf("expected sequence failure in response body, got %s", w.Body.String())
	}
}

func TestCreateStockTransferFailsWhenSequenceIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sourceBranchID := primitive.NewObjectID()
	targetBranchID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	stockID := primitive.NewObjectID()

	repo := &stockTransferRepoStub{
		createFn: func(form request.StockTransfer) (*entities.StockTransfer, error) {
			t.Fatal("create stock transfer should not be called when sequence is missing")
			return nil, nil
		},
	}
	productRepo := &transferProductStub{}
	sequenceRepo := &transferSequenceStub{
		nextFn: func(field string) (*entities.Sequence, error) {
			return nil, nil
		},
	}

	body := `{"toBranchId":"` + targetBranchID.Hex() + `","items":[{"productId":"` + productID.Hex() + `","stockId":"` + stockID.Hex() + `","quantity":2}]}`
	req := httptest.NewRequest(http.MethodPost, "/stock-transfers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", sourceBranchID.Hex())

	CreateStockTransfer(repo, productRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "stock transfer sequence not available") {
		t.Fatalf("expected missing sequence error in response body, got %s", w.Body.String())
	}
}

func TestApproveStockTransferReturnsErrorWhenTransactionalApproveFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	sourceBranchID := primitive.NewObjectID()
	transferID := primitive.NewObjectID()

	repo := &stockTransferRepoStub{
		getByIDFn: func(id string) (*entities.StockTransfer, error) {
			return &entities.StockTransfer{
				Id:           transferID,
				FromBranchId: sourceBranchID,
				ToBranchId:   primitive.NewObjectID(),
				Status:       "PENDING",
			}, nil
		},
		approveFn: func(id string, updatedBy string) (*entities.StockTransfer, error) {
			return nil, errors.New("status update failed")
		},
	}
	productRepo := &transferProductStub{}

	req := httptest.NewRequest(http.MethodPatch, "/stock-transfers/"+transferID.Hex()+"/approve", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: transferID.Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", sourceBranchID.Hex())

	ApproveStockTransfer(repo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.TR_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.TR_BAD_REQUEST_002, w.Body.String())
	}
}
