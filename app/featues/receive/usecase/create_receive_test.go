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

type createReceiveRepoStub struct {
	repositories.IReceive
	createReceiveFn              func(form request.Receive) (*entities.Receive, error)
	createReceiveItemFn          func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error)
	updateReceiveTotalCostByIDFn func(id string, totalCost float64) (*entities.Receive, error)
	removeReceiveByIDFn          func(id string) (*entities.Receive, error)
}

func (s *createReceiveRepoStub) CreateReceive(form request.Receive) (*entities.Receive, error) {
	return s.createReceiveFn(form)
}

func (s *createReceiveRepoStub) CreateReceiveItem(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
	return s.createReceiveItemFn(receiveId, lotId, productId, form)
}

func (s *createReceiveRepoStub) UpdateReceiveTotalCostById(id string, totalCost float64) (*entities.Receive, error) {
	return s.updateReceiveTotalCostByIDFn(id, totalCost)
}

func (s *createReceiveRepoStub) RemoveReceiveById(id string) (*entities.Receive, error) {
	return s.removeReceiveByIDFn(id)
}

type createReceiveProductRepoStub struct {
	repositories.IProduct
	getProductByIDFn func(id string) (*entities.Product, error)
}

func (s *createReceiveProductRepoStub) GetProductById(id string) (*entities.Product, error) {
	return s.getProductByIDFn(id)
}

type receiveSequenceRepoStub struct {
	repositories.ISequence
	nextSequenceFn func(name string) (*entities.Sequence, error)
}

func (s *receiveSequenceRepoStub) NextSequence(name string) (*entities.Sequence, error) {
	return s.nextSequenceFn(name)
}

func TestCreateReceiveFailsWhenSequenceLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveRepo := &createReceiveRepoStub{
		createReceiveFn: func(form request.Receive) (*entities.Receive, error) {
			t.Fatal("create receive should not be called when sequence fails")
			return nil, nil
		},
		createReceiveItemFn: func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
			t.Fatal("create receive item should not be called when sequence fails")
			return nil, nil
		},
		updateReceiveTotalCostByIDFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called when sequence fails")
			return nil, nil
		},
		removeReceiveByIDFn: func(id string) (*entities.Receive, error) {
			t.Fatal("remove receive should not be called when sequence fails")
			return nil, nil
		},
	}
	productRepo := &createReceiveProductRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, nil
		},
	}
	sequenceRepo := &receiveSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return nil, errors.New("sequence failed")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/receives", strings.NewReader(`{"supplierId":"`+primitive.NewObjectID().Hex()+`","items":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateReceive(receiveRepo, sequenceRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.RC_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s, got %s", errcode.RC_BAD_REQUEST_002, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "sequence failed") {
		t.Fatalf("expected sequence failure in response, got %s", w.Body.String())
	}
}

func TestCreateReceiveFailsWhenSequenceIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveRepo := &createReceiveRepoStub{
		createReceiveFn: func(form request.Receive) (*entities.Receive, error) {
			t.Fatal("create receive should not be called when sequence is missing")
			return nil, nil
		},
		createReceiveItemFn: func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
			t.Fatal("create receive item should not be called when sequence is missing")
			return nil, nil
		},
		updateReceiveTotalCostByIDFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called when sequence is missing")
			return nil, nil
		},
		removeReceiveByIDFn: func(id string) (*entities.Receive, error) {
			t.Fatal("remove receive should not be called when sequence is missing")
			return nil, nil
		},
	}
	productRepo := &createReceiveProductRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, nil
		},
	}
	sequenceRepo := &receiveSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) { return nil, nil },
	}

	req := httptest.NewRequest(http.MethodPost, "/receives", strings.NewReader(`{"supplierId":"`+primitive.NewObjectID().Hex()+`","items":[]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateReceive(receiveRepo, sequenceRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "receive sequence not available") {
		t.Fatalf("expected missing sequence error, got %s", w.Body.String())
	}
}

func TestCreateReceiveRollsBackWhenCreateReceiveItemFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	var removedReceiveID string

	receiveRepo := &createReceiveRepoStub{
		createReceiveFn: func(form request.Receive) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID}, nil
		},
		createReceiveItemFn: func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
			return nil, errors.New("create receive item failed")
		},
		updateReceiveTotalCostByIDFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called when create receive item fails")
			return nil, nil
		},
		removeReceiveByIDFn: func(id string) (*entities.Receive, error) {
			removedReceiveID = id
			return &entities.Receive{Id: receiveID}, nil
		},
	}
	productRepo := &createReceiveProductRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return &entities.Product{
				Id:           productID,
				Name:         "Drug A",
				SerialNumber: "SN001",
				Price:        10,
				Unit:         "TAB",
			}, nil
		},
	}
	sequenceRepo := &receiveSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: name, Prefix: "RC", Value: 1, Format: 4}, nil
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","items":[{"productId":"` + productID.Hex() + `","costPrice":5,"quantity":2,"lotNumber":"LOT-1"}]}`
	req := httptest.NewRequest(http.MethodPost, "/receives", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateReceive(receiveRepo, sequenceRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if removedReceiveID != receiveID.Hex() {
		t.Fatalf("expected rollback for receive %s, got %s", receiveID.Hex(), removedReceiveID)
	}
	if !strings.Contains(w.Body.String(), "create receive item failed") {
		t.Fatalf("expected create receive item error, got %s", w.Body.String())
	}
}

func TestCreateReceiveRollsBackWhenTotalCostUpdateFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	var removedReceiveID string

	receiveRepo := &createReceiveRepoStub{
		createReceiveFn: func(form request.Receive) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID}, nil
		},
		createReceiveItemFn: func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
			return &entities.ReceiveItem{}, nil
		},
		updateReceiveTotalCostByIDFn: func(id string, totalCost float64) (*entities.Receive, error) {
			return nil, errors.New("update total cost failed")
		},
		removeReceiveByIDFn: func(id string) (*entities.Receive, error) {
			removedReceiveID = id
			return &entities.Receive{Id: receiveID}, nil
		},
	}
	productRepo := &createReceiveProductRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return &entities.Product{
				Id:           productID,
				Name:         "Drug A",
				SerialNumber: "SN001",
				Price:        10,
				Unit:         "TAB",
			}, nil
		},
	}
	sequenceRepo := &receiveSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: name, Prefix: "RC", Value: 1, Format: 4}, nil
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","items":[{"productId":"` + productID.Hex() + `","costPrice":5,"quantity":2,"lotNumber":"LOT-1"}]}`
	req := httptest.NewRequest(http.MethodPost, "/receives", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateReceive(receiveRepo, sequenceRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if removedReceiveID != receiveID.Hex() {
		t.Fatalf("expected rollback for receive %s, got %s", receiveID.Hex(), removedReceiveID)
	}
	if !strings.Contains(w.Body.String(), "update total cost failed") {
		t.Fatalf("expected total cost update error, got %s", w.Body.String())
	}
}

func TestCreateReceiveFailsWhenExpireDateIsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveRepo := &createReceiveRepoStub{
		createReceiveFn: func(form request.Receive) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID()}, nil
		},
		createReceiveItemFn: func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
			t.Fatal("create receive item should not be called when expire date is invalid")
			return nil, nil
		},
		updateReceiveTotalCostByIDFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called when expire date is invalid")
			return nil, nil
		},
		removeReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID()}, nil
		},
	}
	productRepo := &createReceiveProductRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return &entities.Product{
				Id:           primitive.NewObjectID(),
				Name:         "Drug A",
				SerialNumber: "SN001",
				Price:        10,
				Unit:         "TAB",
			}, nil
		},
	}
	sequenceRepo := &receiveSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: name, Prefix: "RC", Value: 1, Format: 4}, nil
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","items":[{"productId":"` + primitive.NewObjectID().Hex() + `","costPrice":5,"quantity":2,"expireDate":"31/12/2026"}]}`
	req := httptest.NewRequest(http.MethodPost, "/receives", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateReceive(receiveRepo, sequenceRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid expireDate") {
		t.Fatalf("expected invalid expireDate error, got %s", w.Body.String())
	}
}

func TestCreateReceiveRollsBackWhenProductLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	var removedReceiveID string

	receiveRepo := &createReceiveRepoStub{
		createReceiveFn: func(form request.Receive) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID}, nil
		},
		createReceiveItemFn: func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
			t.Fatal("create receive item should not be called when product lookup fails")
			return nil, nil
		},
		updateReceiveTotalCostByIDFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called when product lookup fails")
			return nil, nil
		},
		removeReceiveByIDFn: func(id string) (*entities.Receive, error) {
			removedReceiveID = id
			return &entities.Receive{Id: receiveID}, nil
		},
	}
	productRepo := &createReceiveProductRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, errors.New("product lookup failed")
		},
	}
	sequenceRepo := &receiveSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: name, Prefix: "RC", Value: 1, Format: 4}, nil
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","items":[{"productId":"` + productID.Hex() + `","costPrice":5,"quantity":2,"lotNumber":"LOT-1"}]}`
	req := httptest.NewRequest(http.MethodPost, "/receives", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateReceive(receiveRepo, sequenceRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if removedReceiveID != receiveID.Hex() {
		t.Fatalf("expected rollback for receive %s, got %s", receiveID.Hex(), removedReceiveID)
	}
	if !strings.Contains(w.Body.String(), "product lookup failed") {
		t.Fatalf("expected product lookup error, got %s", w.Body.String())
	}
}

func TestCreateReceiveRollsBackWhenProductNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	productID := primitive.NewObjectID()
	var removedReceiveID string

	receiveRepo := &createReceiveRepoStub{
		createReceiveFn: func(form request.Receive) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID}, nil
		},
		createReceiveItemFn: func(receiveId string, lotId string, productId string, form request.Product) (*entities.ReceiveItem, error) {
			t.Fatal("create receive item should not be called when product is missing")
			return nil, nil
		},
		updateReceiveTotalCostByIDFn: func(id string, totalCost float64) (*entities.Receive, error) {
			t.Fatal("update total cost should not be called when product is missing")
			return nil, nil
		},
		removeReceiveByIDFn: func(id string) (*entities.Receive, error) {
			removedReceiveID = id
			return &entities.Receive{Id: receiveID}, nil
		},
	}
	productRepo := &createReceiveProductRepoStub{
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, nil
		},
	}
	sequenceRepo := &receiveSequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return &entities.Sequence{Field: name, Prefix: "RC", Value: 1, Format: 4}, nil
		},
	}

	body := `{"supplierId":"` + primitive.NewObjectID().Hex() + `","items":[{"productId":"` + productID.Hex() + `","costPrice":5,"quantity":2,"lotNumber":"LOT-1"}]}`
	req := httptest.NewRequest(http.MethodPost, "/receives", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	CreateReceive(receiveRepo, sequenceRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if removedReceiveID != receiveID.Hex() {
		t.Fatalf("expected rollback for receive %s, got %s", receiveID.Hex(), removedReceiveID)
	}
	if !strings.Contains(w.Body.String(), "product "+productID.Hex()+" not found") {
		t.Fatalf("expected product not found error, got %s", w.Body.String())
	}
}
