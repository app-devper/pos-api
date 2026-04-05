package usecase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type promptPaySettingStub struct {
	repositories.ISetting
	getSettingByBranchIDFn func(branchId string) (*entities.Setting, error)
}

func (s *promptPaySettingStub) GetSettingByBranchId(branchId string) (*entities.Setting, error) {
	return s.getSettingByBranchIDFn(branchId)
}

func TestGetPromptPayPayloadRejectsInvalidAmount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	settingRepo := &promptPaySettingStub{
		getSettingByBranchIDFn: func(branchId string) (*entities.Setting, error) {
			return &entities.Setting{PromptPayId: "0812345678"}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/promptpay/payload?amount=abc", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetPromptPayPayload(settingRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "amount is not valid") {
		t.Fatalf("expected invalid amount error, got %s", w.Body.String())
	}
}

func TestGetPromptPayPayloadFailsWhenSettingLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	settingRepo := &promptPaySettingStub{
		getSettingByBranchIDFn: func(branchId string) (*entities.Setting, error) {
			return nil, errors.New("setting lookup failed")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/promptpay/payload?amount=10", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", primitive.NewObjectID().Hex())

	GetPromptPayPayload(settingRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "setting lookup failed") {
		t.Fatalf("expected setting lookup failure, got %s", w.Body.String())
	}
}
