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
	createPromotionFn func(form request.Promotion) (*entities.Promotion, error)
}

func (s *promotionRepoStub) CreatePromotion(form request.Promotion) (*entities.Promotion, error) {
	return s.createPromotionFn(form)
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
