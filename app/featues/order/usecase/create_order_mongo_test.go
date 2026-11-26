package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/db"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// orderCodePattern is an Order's code: OD_, the day as yyyymmdd and the
// day's four-digit sequence.
var orderCodePattern = regexp.MustCompile(`^OD_\d{12}$`)

// newSaleDatabase is a throwaway POS database on MONGO_TEST_URI with one
// Product sold by the tablet at 10; the test is skipped without one.
func newSaleDatabase(t *testing.T) (*db.Resource, string) {
	t.Helper()
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set")
	}
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatal(err)
	}
	pos := client.Database(fmt.Sprintf("pos_create_order_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	product, unit := primitive.NewObjectID(), primitive.NewObjectID()
	docs := map[string]any{
		"products":       entities.Product{Id: product, Name: "Paracetamol", Unit: "TAB"},
		"product_units":  entities.ProductUnit{Id: unit, ProductId: product, Unit: "TAB", CostPrice: 3},
		"product_prices": entities.ProductPrice{Id: primitive.NewObjectID(), ProductId: product, UnitId: unit, CustomerType: "General", Price: 10},
	}
	for collection, doc := range docs {
		if _, err := pos.Collection(collection).InsertOne(ctx, doc); err != nil {
			t.Fatal(err)
		}
	}
	body := fmt.Sprintf(`{"saleId":"s1","type":"CASH","payments":[{"amount":50,"type":"CASH"}],"items":[{"productId":%q,"unitId":%q,"quantity":2,"priceType":"General","allowOversell":true}]}`, product.Hex(), unit.Hex())
	return &db.Resource{Client: client, PosDb: pos}, body
}

type createdOrder struct {
	Data struct {
		Id   string `json:"id"`
		Code string `json:"code"`
	} `json:"data"`
}

func postRecordedSale(t *testing.T, handler gin.HandlerFunc, branch string, body string) createdOrder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/orders", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Set("UserId", "user-1")
	ctx.Set("BranchId", branch)
	handler(ctx)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d %s", w.Code, w.Body.String())
	}
	var got createdOrder
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response %s: %v", w.Body.String(), err)
	}
	return got
}

func TestCreateOrderAnswersTheOrderCodeAndId(t *testing.T) {
	resource, body := newSaleDatabase(t)
	handler := CreateOrder(repositories.NewLedger(resource))
	branch := primitive.NewObjectID().Hex()

	first := postRecordedSale(t, handler, branch, body)
	again := postRecordedSale(t, handler, branch, body)

	if !orderCodePattern.MatchString(first.Data.Code) {
		t.Fatalf("code %q does not match %s", first.Data.Code, orderCodePattern)
	}
	if id, err := primitive.ObjectIDFromHex(first.Data.Id); err != nil || id.IsZero() {
		t.Fatalf("id %q: %v", first.Data.Id, err)
	}
	if again.Data.Code != first.Data.Code || again.Data.Id != first.Data.Id {
		t.Fatalf("repeat answered %+v, first %+v", again.Data, first.Data)
	}
}
