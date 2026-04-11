package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

type promotionRepoStub struct {
	repositories.IPromotion
	createPromotionFn  func(form request.Promotion) (*entities.Promotion, error)
	getPromotionByIDFn func(id string, branchId string) (*entities.Promotion, error)
	updatePromotionFn  func(id string, branchId string, form request.UpdatePromotion) (*entities.Promotion, error)
	removePromotionFn  func(id string, branchId string, updatedBy string) (*entities.Promotion, error)
}

func (s *promotionRepoStub) CreatePromotion(form request.Promotion) (*entities.Promotion, error) {
	return s.createPromotionFn(form)
}

func (s *promotionRepoStub) GetPromotionById(id string, branchId string) (*entities.Promotion, error) {
	return s.getPromotionByIDFn(id, branchId)
}

func (s *promotionRepoStub) UpdatePromotionById(id string, branchId string, form request.UpdatePromotion) (*entities.Promotion, error) {
	return s.updatePromotionFn(id, branchId, form)
}

func (s *promotionRepoStub) RemovePromotionById(id string, branchId string, updatedBy string) (*entities.Promotion, error) {
	return s.removePromotionFn(id, branchId, updatedBy)
}

func TestCreatePromotionPassesBranchAndCreatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	branchID := "507f1f77bcf86cd799439011"
	var gotForm request.Promotion

	repo := &promotionRepoStub{
		createPromotionFn: func(form request.Promotion) (*entities.Promotion, error) {
			gotForm = form
			return &entities.Promotion{Code: form.Code, Name: form.Name}, nil
		},
	}

	body := `{"code":"PM001","name":"Promo A","type":"PERCENT","value":10,"startDate":"` + startDate.Format(time.RFC3339) + `","endDate":"` + endDate.Format(time.RFC3339) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/promotions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branchID)

	CreatePromotion(repo)(ctx)

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

func TestGetPromotionByIdPassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promotionID := "507f1f77bcf86cd799439011"
	branchID := "507f1f77bcf86cd799439012"
	var gotID string
	var gotBranchID string
	repo := &promotionRepoStub{
		getPromotionByIDFn: func(id string, branchId string) (*entities.Promotion, error) {
			gotID = id
			gotBranchID = branchId
			return &entities.Promotion{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/promotions/"+promotionID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: promotionID}}
	ctx.Set("BranchId", branchID)

	GetPromotionById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != promotionID {
		t.Fatalf("expected promotion id %s, got %s", promotionID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
}

func TestUpdatePromotionByIdPassesBranchIdAndUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promotionID := "507f1f77bcf86cd799439011"
	branchID := "507f1f77bcf86cd799439012"
	var gotID string
	var gotBranchID string
	var gotForm request.UpdatePromotion
	repo := &promotionRepoStub{
		updatePromotionFn: func(id string, branchId string, form request.UpdatePromotion) (*entities.Promotion, error) {
			gotID = id
			gotBranchID = branchId
			gotForm = form
			return &entities.Promotion{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/promotions/"+promotionID, strings.NewReader(`{"name":"Promo B"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: promotionID}}
	ctx.Set("BranchId", branchID)
	ctx.Set("UserId", "admin-1")

	UpdatePromotionById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != promotionID {
		t.Fatalf("expected promotion id %s, got %s", promotionID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
	if gotForm.UpdatedBy != "admin-1" {
		t.Fatalf("expected UpdatedBy admin-1, got %s", gotForm.UpdatedBy)
	}
}

func TestDeletePromotionByIdPassesBranchIdAndUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	promotionID := "507f1f77bcf86cd799439011"
	branchID := "507f1f77bcf86cd799439012"
	var gotID string
	var gotBranchID string
	var gotUpdatedBy string
	repo := &promotionRepoStub{
		removePromotionFn: func(id string, branchId string, updatedBy string) (*entities.Promotion, error) {
			gotID = id
			gotBranchID = branchId
			gotUpdatedBy = updatedBy
			return &entities.Promotion{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/promotions/"+promotionID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: promotionID}}
	ctx.Set("BranchId", branchID)
	ctx.Set("UserId", "admin-1")

	DeletePromotionById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != promotionID {
		t.Fatalf("expected promotion id %s, got %s", promotionID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
	if gotUpdatedBy != "admin-1" {
		t.Fatalf("expected UpdatedBy admin-1, got %s", gotUpdatedBy)
	}
}
