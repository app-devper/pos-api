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

type barcodeProductStub struct {
	repositories.IProduct
	getProductsByIDsFn func(ids []string) ([]entities.Product, error)
	getProductByIDFn   func(id string) (*entities.Product, error)
	batchCalls         int
	singleCalls        int
}

func (s *barcodeProductStub) GetProductsByIds(ids []string) ([]entities.Product, error) {
	s.batchCalls++
	return s.getProductsByIDsFn(ids)
}

func (s *barcodeProductStub) GetProductById(id string) (*entities.Product, error) {
	s.singleCalls++
	return s.getProductByIDFn(id)
}

func TestGetBarcodePDFUsesBatchProductLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productID := primitive.NewObjectID()
	productRepo := &barcodeProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return []entities.Product{{
				Id:           productID,
				Name:         "Drug A",
				SerialNumber: "1234567890",
				Price:        10.5,
				Unit:         "BOX",
			}}, nil
		},
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, errors.New("single product lookup should not be used")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/reports/barcode", strings.NewReader(`{"productIds":["`+productID.Hex()+`"],"copies":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	GetBarcodePDF(productRepo)(ctx)

	if productRepo.batchCalls != 1 || productRepo.singleCalls != 0 {
		t.Fatalf("expected batch lookup once and no single lookups, got batch=%d single=%d", productRepo.batchCalls, productRepo.singleCalls)
	}
	if w.Code == http.StatusBadRequest {
		t.Fatalf("expected request to get past product lookup, got %d with body %s", w.Code, w.Body.String())
	}
}

func TestGetBarcodePDFFailsWhenBatchProductLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	productRepo := &barcodeProductStub{
		getProductsByIDsFn: func(ids []string) ([]entities.Product, error) {
			return nil, errors.New("product lookup failed")
		},
		getProductByIDFn: func(id string) (*entities.Product, error) {
			return nil, errors.New("single product lookup should not be used")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/reports/barcode", strings.NewReader(`{"productIds":["`+primitive.NewObjectID().Hex()+`"],"copies":1}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req

	GetBarcodePDF(productRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "product lookup failed") {
		t.Fatalf("expected product lookup failure in response, got %s", w.Body.String())
	}
}
