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

type settingRepoStub struct {
	repositories.ISetting
	getFn    func(branchId string) (*entities.Setting, error)
	upsertFn func(form request.Setting) (*entities.Setting, error)
}

func (s *settingRepoStub) GetSettingByBranchId(branchId string) (*entities.Setting, error) {
	return s.getFn(branchId)
}

func (s *settingRepoStub) UpsertSetting(form request.Setting) (*entities.Setting, error) {
	return s.upsertFn(form)
}

func TestGetSettingPassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	var gotBranchID string
	repo := &settingRepoStub{
		getFn: func(branchId string) (*entities.Setting, error) {
			gotBranchID = branchId
			return &entities.Setting{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", branchID)

	GetSetting(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected BranchId %s, got %s", branchID, gotBranchID)
	}
}

func TestUpsertSettingPassesBranchIdAndUpdatedBy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	branchID := primitive.NewObjectID().Hex()
	var gotForm request.Setting
	repo := &settingRepoStub{
		upsertFn: func(form request.Setting) (*entities.Setting, error) {
			gotForm = form
			return &entities.Setting{CompanyName: form.CompanyName}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(`{"companyName":"HQ","companyAddress":"Bangkok"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("BranchId", branchID)
	ctx.Set("UserId", "user-1")

	UpsertSetting(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotForm.BranchId != branchID {
		t.Fatalf("expected BranchId %s, got %s", branchID, gotForm.BranchId)
	}
	if gotForm.UpdatedBy != "user-1" {
		t.Fatalf("expected UpdatedBy user-1, got %s", gotForm.UpdatedBy)
	}
}
