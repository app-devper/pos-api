package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/ledger"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type productLedgerStub struct {
	deleteFn func(productID, by string) (*entities.Product, error)
}

func (s productLedgerStub) DeleteProduct(_ context.Context, productID, by string) (*entities.Product, error) {
	return s.deleteFn(productID, by)
}

func deleteProduct(t *testing.T, stub productLedgerStub, id string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodDelete, "/products/"+id, nil)
	ctx.Params = gin.Params{{Key: "productId", Value: id}}
	ctx.Set("UserId", "user-1")
	DeleteProductById(stub)(ctx)
	return w
}

func TestDeleteProductByIdGoesThroughTheLedgerAndAnswersTheProduct(t *testing.T) {
	id := primitive.NewObjectID()
	now := time.Now()
	var gotID, gotBy string
	stub := productLedgerStub{deleteFn: func(productID, by string) (*entities.Product, error) {
		gotID, gotBy = productID, by
		return &entities.Product{Id: id, Name: "Paracetamol", DeletedDate: &now}, nil
	}}

	w := deleteProduct(t, stub, id.Hex())

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusOK || body["id"] != id.Hex() || body["name"] != "Paracetamol" || body["deletedDate"] == nil {
		t.Fatalf("got %d %s, want 200 with the deleted product", w.Code, w.Body.String())
	}
	if gotID != id.Hex() || gotBy != "user-1" {
		t.Fatalf("ledger got id=%q by=%q", gotID, gotBy)
	}
}

func TestDeleteProductByIdMapsLedgerErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		code int
		want string
	}{
		{"already deleted", fmt.Errorf("%w: product x", ledger.ErrNotFound), http.StatusNotFound, errcode.SY_NOT_FOUND_002},
		{"bad id", &ledger.Rejected{Reason: "invalid product id"}, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002},
		{"conflict", ledger.ErrConflict, http.StatusConflict, errcode.SY_CONFLICT_001},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			stub := productLedgerStub{deleteFn: func(string, string) (*entities.Product, error) { return nil, c.err }}

			w := deleteProduct(t, stub, "x")

			var body struct {
				Code string `json:"errcode"`
			}
			_ = json.Unmarshal(w.Body.Bytes(), &body)
			if w.Code != c.code || body.Code != c.want {
				t.Fatalf("got %d %s, want %d %s", w.Code, w.Body.String(), c.code, c.want)
			}
		})
	}
}
