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
	var gotForm request.UpdateReceive

	receiveRepo := &updateReceiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: primitive.NewObjectID(), Code: "RC-001"}, nil
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
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

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
