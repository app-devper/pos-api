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

type allergyProductRepoStub struct {
	repositories.IProduct
	getProductsByIDsFn func(ids []string) ([]entities.Product, error)
}

func (s *allergyProductRepoStub) GetProductsByIds(ids []string) ([]entities.Product, error) {
	return s.getProductsByIDsFn(ids)
}

func TestAllergyCheckReturnsErrorWhenProductLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	patientID := primitive.NewObjectID().Hex()
	branchID := primitive.NewObjectID().Hex()
	patientRepo := &patientRepoStub{
		getByIDFn: func(id string, branchId string) (*entities.Patient, error) {
			return &entities.Patient{
				Id: primitive.NewObjectID(),
				Allergies: []entities.DrugAllergy{
					{DrugName: "amoxicillin"},
				},
			}, nil
		},
	}
	productRepo := &allergyProductRepoStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return nil, errors.New("product lookup failed")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/patients/"+patientID+"/allergy-check", strings.NewReader(`{"productIds":["`+primitive.NewObjectID().Hex()+`"]}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Params = gin.Params{{Key: "id", Value: patientID}}
	ctx.Set("BranchId", branchID)

	AllergyCheck(patientRepo, productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), errcode.PT_BAD_REQUEST_002) {
		t.Fatalf("expected errcode %s, got %s", errcode.PT_BAD_REQUEST_002, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "product lookup failed") {
		t.Fatalf("expected product lookup error in response, got %s", w.Body.String())
	}
}
