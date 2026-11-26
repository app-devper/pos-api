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

type orderRangeRepoStub struct {
	repositories.IOrder
	calls []request.GetOrderRange
}

func (s *orderRangeRepoStub) GetOrderRange(form request.GetOrderRange) ([]entities.Order, error) {
	s.calls = append(s.calls, form)
	return []entities.Order{}, nil
}

const orderRangeQuery = "/orders?startDate=2026-11-01&endDate=2026-11-02"

func serveOrdersRange(repo *orderRangeRepoStub, url string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, url, nil)
	ctx.Set("BranchId", "")
	GetOrdersRange(repo)(ctx)
	return w
}

func TestGetOrdersRangeKeepsCancelledBillsByDefault(t *testing.T) {
	repo := &orderRangeRepoStub{}

	w := serveOrdersRange(repo, orderRangeQuery)

	if w.Code != http.StatusOK || len(repo.calls) != 1 || repo.calls[0].ExcludeCancelled {
		t.Fatalf("status %d calls %+v", w.Code, repo.calls)
	}
}

func TestGetOrdersRangePassesExcludeCancelled(t *testing.T) {
	repo := &orderRangeRepoStub{}

	w := serveOrdersRange(repo, orderRangeQuery+"&excludeCancelled=true")

	if w.Code != http.StatusOK || len(repo.calls) != 1 || !repo.calls[0].ExcludeCancelled {
		t.Fatalf("status %d calls %+v", w.Code, repo.calls)
	}
}

func TestGetOrdersRangeRejectsABadExcludeCancelled(t *testing.T) {
	for name, query := range map[string]string{
		"maybe":    "&excludeCancelled=maybe",
		"empty":    "&excludeCancelled=",
		"repeated": "&excludeCancelled=true&excludeCancelled=maybe",
	} {
		t.Run(name, func(t *testing.T) {
			repo := &orderRangeRepoStub{}

			w := serveOrdersRange(repo, orderRangeQuery+query)

			if w.Code != http.StatusBadRequest || len(repo.calls) != 0 {
				t.Fatalf("status %d calls %+v", w.Code, repo.calls)
			}
			body := w.Body.String()
			if !strings.Contains(body, errcode.OR_BAD_REQUEST_001) || !strings.Contains(body, "excludeCancelled") {
				t.Fatalf("body %s", body)
			}
		})
	}
}
