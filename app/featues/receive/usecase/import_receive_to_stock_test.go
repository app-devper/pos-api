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

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type receiveRepoStub struct {
	repositories.IReceive
	getReceiveByIDFn       func(id string) (*entities.Receive, error)
	importReceiveToStockFn func(receiveId string, userId string, branchId string) (*entities.Receive, error)
}

func (s *receiveRepoStub) GetReceiveById(id string) (*entities.Receive, error) {
	return s.getReceiveByIDFn(id)
}

func (s *receiveRepoStub) ImportReceiveToStock(receiveId string, userId string, branchId string) (*entities.Receive, error) {
	return s.importReceiveToStockFn(receiveId, userId, branchId)
}

func TestImportReceiveToStockReturnsErrorWhenTransactionalImportFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	receiveID := primitive.NewObjectID()
	branchID := primitive.NewObjectID()

	receiveRepo := &receiveRepoStub{
		getReceiveByIDFn: func(id string) (*entities.Receive, error) {
			return &entities.Receive{Id: receiveID, BranchId: branchID}, nil
		},
		importReceiveToStockFn: func(receiveId string, userId string, branchId string) (*entities.Receive, error) {
			return nil, errors.New("status failed")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/receives/"+receiveID.Hex()+"/import", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "receiveId", Value: receiveID.Hex()}}
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID.Hex())

	ImportReceiveToStock(receiveRepo, nil)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.RC_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s in response body, got %s", errcode.RC_BAD_REQUEST_002, w.Body.String())
	}
}
