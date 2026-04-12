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

type patientRepoStub struct {
	repositories.IPatient
	getByIDFn    func(id string, branchId string) (*entities.Patient, error)
	updateByIDFn func(id string, form request.UpdatePatient, branchId string) (*entities.Patient, error)
	removeByIDFn func(id string, branchId string, updatedBy string) (*entities.Patient, error)
}

func (s *patientRepoStub) GetPatientById(id string, branchId string) (*entities.Patient, error) {
	return s.getByIDFn(id, branchId)
}

func (s *patientRepoStub) UpdatePatientById(id string, form request.UpdatePatient, branchId string) (*entities.Patient, error) {
	return s.updateByIDFn(id, form, branchId)
}

func (s *patientRepoStub) RemovePatientById(id string, branchId string, updatedBy string) (*entities.Patient, error) {
	return s.removeByIDFn(id, branchId, updatedBy)
}

func TestGetPatientByIdPassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	var gotID string
	var gotBranchID string

	repo := &patientRepoStub{
		getByIDFn: func(id string, branchId string) (*entities.Patient, error) {
			gotID = id
			gotBranchID = branchId
			return &entities.Patient{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/patients/"+patientID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: patientID}}
	ctx.Set("BranchId", branchID)

	GetPatientById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != patientID {
		t.Fatalf("expected patient id %s, got %s", patientID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
}

func TestUpdatePatientByIdPassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	var gotID string
	var gotBranchID string

	repo := &patientRepoStub{
		updateByIDFn: func(id string, form request.UpdatePatient, branchId string) (*entities.Patient, error) {
			gotID = id
			gotBranchID = branchId
			return &entities.Patient{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/patients/"+patientID, strings.NewReader(`{"firstName":"A"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: patientID}}
	ctx.Set("BranchId", branchID)
	ctx.Set("UserId", "user-1")

	UpdatePatientById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != patientID {
		t.Fatalf("expected patient id %s, got %s", patientID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
}

func TestDeletePatientByIdPassesBranchId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	var gotID string
	var gotBranchID string
	var gotUpdatedBy string

	repo := &patientRepoStub{
		removeByIDFn: func(id string, branchId string, updatedBy string) (*entities.Patient, error) {
			gotID = id
			gotBranchID = branchId
			gotUpdatedBy = updatedBy
			return &entities.Patient{Id: primitive.NewObjectID()}, nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/patients/"+patientID, nil)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: patientID}}
	ctx.Set("BranchId", branchID)
	ctx.Set("UserId", "user-1")

	DeletePatientById(repo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotID != patientID {
		t.Fatalf("expected patient id %s, got %s", patientID, gotID)
	}
	if gotBranchID != branchID {
		t.Fatalf("expected branch id %s, got %s", branchID, gotBranchID)
	}
	if gotUpdatedBy != "user-1" {
		t.Fatalf("expected UpdatedBy user-1, got %s", gotUpdatedBy)
	}
}
