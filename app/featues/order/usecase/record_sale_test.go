package usecase

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type saleRepoStub struct {
	repositories.IOrder
	found    *repositories.RecordedSale
	findErr  error
	record   func(form request.Sale) (*repositories.RecordedSale, error)
	recorded []request.Sale
}

func (s *saleRepoStub) FindSale(form request.Sale) (*repositories.RecordedSale, error) {
	return s.found, s.findErr
}

func (s *saleRepoStub) RecordSale(form request.Sale) (*repositories.RecordedSale, error) {
	s.recorded = append(s.recorded, form)
	return s.record(form)
}

func (s *saleRepoStub) CreateOrder(form request.Order) (*entities.Order, []entities.OrderItem, error) {
	return nil, nil, errors.New("old path taken")
}

type countingSequence struct {
	repositories.ISequence
	calls int
}

func (s *countingSequence) NextSequence(field string) (*entities.Sequence, error) {
	s.calls++
	return &entities.Sequence{Field: field, Value: s.calls, Format: 4}, nil
}

const saleBody = `{"saleId":"s1","type":"CASH","payments":[{"amount":50,"type":"CASH"}],"items":[{"productId":"p","unitId":"u","quantity":2,"priceType":"General"}]}`

func postSale(t *testing.T, repo repositories.IOrder, seq repositories.ISequence, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", primitive.NewObjectID().Hex())
	CreateOrder(repo, &productStub{}, &productStockStub{}, seq)(ctx)
	return w
}

func TestCreateOrderWithSaleIdRecordsTheSale(t *testing.T) {
	order := &entities.Order{Id: primitive.NewObjectID(), Total: 20}
	repo := &saleRepoStub{record: func(form request.Sale) (*repositories.RecordedSale, error) {
		return &repositories.RecordedSale{Order: order, Stocks: []entities.ProductStock{}}, nil
	}}
	seq := &countingSequence{}

	w := postSale(t, repo, seq, saleBody)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data   entities.Order          `json:"data"`
		Stocks []entities.ProductStock `json:"stocks"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil || resp.Data.Id != order.Id || resp.Stocks == nil {
		t.Fatalf("response %s (%v)", w.Body.String(), err)
	}
	if len(repo.recorded) != 1 || seq.calls != 1 {
		t.Fatalf("recorded %d sequences %d", len(repo.recorded), seq.calls)
	}
	s := repo.recorded[0]
	if s.SaleId != "s1" || s.Code == "" || s.CreatedBy != "user-1" || s.BranchId == "" || s.Items[0].PriceType != "General" {
		t.Fatalf("sale %+v", s)
	}
}

func TestCreateOrderRepeatedSaleDoesNotTakeAnOrderCode(t *testing.T) {
	order := &entities.Order{Id: primitive.NewObjectID()}
	repo := &saleRepoStub{found: &repositories.RecordedSale{Order: order}}
	seq := &countingSequence{}

	w := postSale(t, repo, seq, saleBody)

	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), order.Id.Hex()) {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	if seq.calls != 0 || len(repo.recorded) != 0 {
		t.Fatalf("sequences %d recorded %d", seq.calls, len(repo.recorded))
	}
}

func TestCreateOrderSaleErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		findErr, recordErr error
		status             int
		code               string
	}{
		"id reused on find":   {findErr: repositories.ErrSaleConflict, status: http.StatusConflict, code: errcode.OR_CONFLICT_001},
		"id reused on record": {recordErr: repositories.ErrSaleConflict, status: http.StatusConflict, code: errcode.OR_CONFLICT_001},
		"rejected":            {recordErr: &repositories.SaleRejected{Reason: "payment too low"}, status: http.StatusBadRequest, code: errcode.OR_BAD_REQUEST_001},
		"failed":              {recordErr: errors.New("mongo down"), status: http.StatusBadRequest, code: errcode.OR_BAD_REQUEST_002},
	} {
		repo := &saleRepoStub{findErr: tc.findErr, record: func(form request.Sale) (*repositories.RecordedSale, error) {
			return nil, tc.recordErr
		}}
		w := postSale(t, repo, &countingSequence{}, saleBody)
		if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.code) {
			t.Errorf("%s: status %d %s", name, w.Code, w.Body.String())
		}
	}
}

func TestCreateOrderSaleRejectsInvalidBody(t *testing.T) {
	for name, body := range map[string]string{
		"no items":      `{"saleId":"s1","type":"CASH","payments":[{"amount":50,"type":"CASH"}],"items":[]}`,
		"no payments":   `{"saleId":"s1","type":"CASH","items":[{"productId":"p","unitId":"u","quantity":1}]}`,
		"zero quantity": `{"saleId":"s1","type":"CASH","payments":[{"amount":50,"type":"CASH"}],"items":[{"productId":"p","unitId":"u","quantity":0}]}`,
		"not json":      `{`,
	} {
		repo := &saleRepoStub{record: func(form request.Sale) (*repositories.RecordedSale, error) {
			t.Fatalf("%s: recorded", name)
			return nil, nil
		}}
		w := postSale(t, repo, &countingSequence{}, body)
		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), errcode.OR_BAD_REQUEST_001) {
			t.Errorf("%s: status %d %s", name, w.Code, w.Body.String())
		}
	}
}

func TestCreateOrderWithoutSaleIdTakesTheOldPath(t *testing.T) {
	repo := &saleRepoStub{record: func(form request.Sale) (*repositories.RecordedSale, error) {
		t.Fatal("new path taken")
		return nil, nil
	}}
	w := postSale(t, repo, &countingSequence{}, `{"items":[],"amount":20,"type":"cash","total":20}`)
	if !strings.Contains(w.Body.String(), "old path taken") {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
}
