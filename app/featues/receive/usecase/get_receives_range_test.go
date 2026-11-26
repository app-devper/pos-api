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

type receiveRangeRepoStub struct {
	repositories.IReceive
	calls []request.GetReceiveRange
}

func (s *receiveRangeRepoStub) GetReceives(form request.GetReceiveRange) ([]entities.Receive, error) {
	s.calls = append(s.calls, form)
	return []entities.Receive{}, nil
}

const receiveRangeQuery = "/receives?startDate=2026-11-01&endDate=2026-11-02"

func serveReceivesRange(repo *receiveRangeRepoStub, url string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodGet, url, nil)
	ctx.Set("BranchId", "")
	GetReceivesRange(repo)(ctx)
	return w
}

func TestGetReceivesRangeKeepsCancelledReceiptsByDefault(t *testing.T) {
	repo := &receiveRangeRepoStub{}

	w := serveReceivesRange(repo, receiveRangeQuery)

	if w.Code != http.StatusOK || len(repo.calls) != 1 || repo.calls[0].ExcludeCancelled {
		t.Fatalf("status %d calls %+v", w.Code, repo.calls)
	}
}

func TestGetReceivesRangePassesExcludeCancelled(t *testing.T) {
	repo := &receiveRangeRepoStub{}

	w := serveReceivesRange(repo, receiveRangeQuery+"&excludeCancelled=true")

	if w.Code != http.StatusOK || len(repo.calls) != 1 || !repo.calls[0].ExcludeCancelled {
		t.Fatalf("status %d calls %+v", w.Code, repo.calls)
	}
}

func TestGetReceivesRangeRejectsABadExcludeCancelled(t *testing.T) {
	for name, query := range map[string]string{
		"maybe":    "&excludeCancelled=maybe",
		"empty":    "&excludeCancelled=",
		"repeated": "&excludeCancelled=true&excludeCancelled=maybe",
	} {
		t.Run(name, func(t *testing.T) {
			repo := &receiveRangeRepoStub{}

			w := serveReceivesRange(repo, receiveRangeQuery+query)

			if w.Code != http.StatusBadRequest || len(repo.calls) != 0 {
				t.Fatalf("status %d calls %+v", w.Code, repo.calls)
			}
			body := w.Body.String()
			if !strings.Contains(body, errcode.RC_BAD_REQUEST_001) || !strings.Contains(body, "excludeCancelled") {
				t.Fatalf("body %s", body)
			}
		})
	}
}
