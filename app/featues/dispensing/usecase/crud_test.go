package usecase

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type dispensingRepoStub struct {
	repositories.IDispensingLog
	getByIDFn        func(id string, branchId string) (*entities.DispensingLog, error)
	getByPatientIDFn func(patientId string, branchId string) ([]entities.DispensingLog, error)
}

func (s *dispensingRepoStub) GetDispensingLogById(id string, branchId string) (*entities.DispensingLog, error) {
	return s.getByIDFn(id, branchId)
}

func (s *dispensingRepoStub) GetDispensingLogsByPatientId(patientId string, branchId string) ([]entities.DispensingLog, error) {
	return s.getByPatientIDFn(patientId, branchId)
}

func TestGetDispensingLogByIdPassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	var gotID string
	var gotBranchID string

	repo := &dispensingRepoStub{
		getByIDFn: func(id string, branchId string) (*entities.DispensingLog, error) {
			gotID = id
			gotBranchID = branchId
			return &entities.DispensingLog{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/dispensing-logs/"+logID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: logID}}
	ctx.Set("BranchId", branchID)

	GetDispensingLogById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != logID {
		t.Fatalf("expected log id %s, got %s", logID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
}

func TestGetDispensingLogsByPatientIdPassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	var gotPatientID string
	var gotBranchID string

	repo := &dispensingRepoStub{
		getByPatientIDFn: func(id string, branchId string) ([]entities.DispensingLog, error) {
			gotPatientID = id
			gotBranchID = branchId
			return []entities.DispensingLog{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/dispensing-logs/patient/"+patientID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "patientId", Value: patientID}}
	ctx.Set("BranchId", branchID)

	GetDispensingLogsByPatientId(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotPatientID != patientID {
		t.Fatalf("expected patient id %s, got %s", patientID, gotPatientID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
}
