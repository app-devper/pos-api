package usecase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

type customerListRepoStub struct {
	repositories.ICustomer
	calls []request.GetCustomers
}

func (s *customerListRepoStub) GetCustomerAll(form request.GetCustomers) ([]entities.Customer, error) {
	s.calls = append(s.calls, form)
	return []entities.Customer{}, nil
}

func serveCustomers(repo *customerListRepoStub, url string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, url, nil)
	GetCustomers(repo)(ctx)
	return w
}

func TestGetCustomersKeepsArchivedCustomersByDefault(t *testing.T) {
	repo := &customerListRepoStub{}

	w := serveCustomers(repo, "/customers")

	if w.Code != http.StatusOK || len(repo.calls) != 1 || repo.calls[0].ExcludeArchived {
		t.Fatalf("status %d calls %+v", w.Code, repo.calls)
	}
}

func TestGetCustomersPassesExcludeArchived(t *testing.T) {
	repo := &customerListRepoStub{}

	w := serveCustomers(repo, "/customers?excludeArchived=true")

	if w.Code != http.StatusOK || len(repo.calls) != 1 || !repo.calls[0].ExcludeArchived {
		t.Fatalf("status %d calls %+v", w.Code, repo.calls)
	}
}

func TestGetCustomersRejectsABadExcludeArchived(t *testing.T) {
	for name, query := range map[string]string{
		"maybe":    "?excludeArchived=maybe",
		"empty":    "?excludeArchived=",
		"repeated": "?excludeArchived=true&excludeArchived=false",
	} {
		t.Run(name, func(t *testing.T) {
			repo := &customerListRepoStub{}

			w := serveCustomers(repo, "/customers"+query)

			if w.Code != http.StatusBadRequest || len(repo.calls) != 0 {
				t.Fatalf("status %d calls %+v", w.Code, repo.calls)
			}
			body := w.Body.String()
			if !strings.Contains(body, errcode.CU_BAD_REQUEST_001) || !strings.Contains(body, "excludeArchived") {
				t.Fatalf("body %s", body)
			}
		})
	}
}
