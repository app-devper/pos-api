package usecase

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

type customerRepoStub struct {
	repositories.ICustomer
	createCustomerFn func(form request.Customer) (*entities.Customer, error)
}

func (s *customerRepoStub) CreateCustomer(form request.Customer) (*entities.Customer, error) {
	return s.createCustomerFn(form)
}

type sequenceRepoStub struct {
	repositories.ISequence
	nextSequenceFn func(name string) (*entities.Sequence, error)
}

func (s *sequenceRepoStub) NextSequence(name string) (*entities.Sequence, error) {
	return s.nextSequenceFn(name)
}

func TestCreateCustomerFailsWhenSequenceLookupFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customerRepo := &customerRepoStub{
		createCustomerFn: func(form request.Customer) (*entities.Customer, error) {
			t.Fatal("create customer should not be called when sequence fails")
			return nil, nil
		},
	}
	sequenceRepo := &sequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return nil, errors.New("sequence failed")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{"name":"Customer A"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")

	CreateCustomer(customerRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "sequence failed") {
		t.Fatalf("expected sequence failure in response, got %s", w.Body.String())
	}
}

func TestCreateCustomerFailsWhenSequenceIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	customerRepo := &customerRepoStub{
		createCustomerFn: func(form request.Customer) (*entities.Customer, error) {
			t.Fatal("create customer should not be called when sequence is missing")
			return nil, nil
		},
	}
	sequenceRepo := &sequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{"name":"Customer A"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")

	CreateCustomer(customerRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
	if !strings.Contains(w.Body.String(), "customer sequence not available") {
		t.Fatalf("expected missing sequence error, got %s", w.Body.String())
	}
}

func TestCreateCustomerUsesGeneratedCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotCustomer request.Customer
	customerRepo := &customerRepoStub{
		createCustomerFn: func(form request.Customer) (*entities.Customer, error) {
			gotCustomer = form
			return &entities.Customer{Code: form.Code, Name: form.Name}, nil
		},
	}
	sequenceRepo := &sequenceRepoStub{
		nextSequenceFn: func(name string) (*entities.Sequence, error) {
			return &entities.Sequence{
				Field:  name,
				Prefix: "MBR",
				Value:  12,
				Format: 4,
			}, nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(`{"name":"Customer A"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = req
	ctx.Set("UserId", "user-1")

	CreateCustomer(customerRepo, sequenceRepo)(ctx)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if gotCustomer.Code == "" {
		t.Fatal("expected generated customer code")
	}
	if gotCustomer.CreatedBy != "user-1" {
		t.Fatalf("expected CreatedBy user-1, got %s", gotCustomer.CreatedBy)
	}
}
