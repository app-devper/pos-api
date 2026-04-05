package usecase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type drugLabelDispensingStub struct {
	repositories.IDispensingLog
	getDispensingLogByIDFn func(id string, branchId string) (*entities.DispensingLog, error)
}

func (s *drugLabelDispensingStub) GetDispensingLogById(id string, branchId string) (*entities.DispensingLog, error) {
	return s.getDispensingLogByIDFn(id, branchId)
}

type drugLabelSettingStub struct {
	repositories.ISetting
	getSettingByBranchIDFn func(branchId string) (*entities.Setting, error)
}

func (s *drugLabelSettingStub) GetSettingByBranchId(branchId string) (*entities.Setting, error) {
	return s.getSettingByBranchIDFn(branchId)
}

func TestGetDrugLabelPDFFailsWhenSettingLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	dispensingRepo := &drugLabelDispensingStub{
		getDispensingLogByIDFn: func(id string, branchId string) (*entities.DispensingLog, error) {
			return &entities.DispensingLog{
				Id:             primitive.NewObjectID(),
				PharmacistName: "เภสัชกร A",
				LicenseNo:      "LIC-001",
				CreatedDate:    time.Now(),
				Items: []entities.DispensingItem{
					{ProductName: "Drug A", Quantity: 1, Unit: "TAB"},
				},
			}, nil
		},
	}
	settingRepo := &drugLabelSettingStub{
		getSettingByBranchIDFn: func(branchId string) (*entities.Setting, error) {
			return nil, errors.New("setting lookup failed")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/reports/drug-label/"+logID+"/pdf", nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "logId", Value: logID}}
	ctx.Set("BranchId", branchID)

	GetDrugLabelPDF(dispensingRepo, settingRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "setting lookup failed") {
		t.Fatalf("expected setting lookup failure, got %s", w.Body.String())
	}
}
