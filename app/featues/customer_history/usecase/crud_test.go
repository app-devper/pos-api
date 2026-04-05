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

type customerHistoryRepoStub struct {
	repositories.ICustomerHistory
	createFn func(form request.CustomerHistory) (*entities.CustomerHistory, error)
	listFn   func(customerCode string, branchId string) ([]entities.CustomerHistory, error)
}

func (s *customerHistoryRepoStub) CreateCustomerHistory(form request.CustomerHistory) (*entities.CustomerHistory, error) {
	return s.createFn(form)
}

func (s *customerHistoryRepoStub) GetCustomerHistories(customerCode string, branchId string) ([]entities.CustomerHistory, error) {
	return s.listFn(customerCode, branchId)
}

func TestCreateCustomerHistoryPassesBranchIdAndCreatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	var gotForm request.CustomerHistory
	repo := &customerHistoryRepoStub{
		createFn: func(form request.CustomerHistory) (*entities.CustomerHistory, error) {
			gotForm = form
			return &entities.CustomerHistory{CustomerCode: form.CustomerCode, Type: form.Type}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/customer-histories", strings.NewReader(`{"customerCode":"C001","type":"SALE","description":"desc"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreateCustomerHistory(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotForm.CreatedBy != "user-1" {
		t.Fatalf("expected CreatedBy user-1, got %s", gotForm.CreatedBy)
	}
	if gotForm.BranchId != branchID {
		t.Fatalf("expected BranchId %s, got %s", branchID, gotForm.BranchId)
	}
}

func TestGetCustomerHistoriesPassesCustomerCodeAndBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	var gotCustomerCode string
	var gotBranchID string
	repo := &customerHistoryRepoStub{
		listFn: func(customerCode string, branchId string) ([]entities.CustomerHistory, error) {
			gotCustomerCode = customerCode
			gotBranchID = branchId
			return []entities.CustomerHistory{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/customer-histories/C001", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "customerCode", Value: "C001"}}
	ctx.Set("BranchId", branchID)

	GetCustomerHistories(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotCustomerCode != "C001" {
		t.Fatalf("expected customer code C001, got %s", gotCustomerCode)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected BranchId %s, got %s", branchID, gotBranchID)
	}
}
