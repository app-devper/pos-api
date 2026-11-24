package legacy

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	bulkOrders        = 20000
	bulkBatch         = 2000
	discountTimeLimit = 15 * time.Second
)

// discountOrder inserts an Order with total and two lines of price 100 (the
// Line's whole amount), quantity 2: the first with discount, the second none.
// Per unit: 90 + 100 = 190; per line: 95 + 100 = 195; Σ price: 200.
func discountOrder(t *testing.T, db *mongo.Database, total, discount float64) primitive.ObjectID {
	t.Helper()
	id := insert(t, db, "orders", bson.M{"total": total})
	insert(t, db, "order_items", bson.M{"orderId": id, "price": 100.0, "quantity": 2, "discount": discount})
	insert(t, db, "order_items", bson.M{"orderId": id, "price": 100.0, "quantity": 2, "discount": 0.0})
	return id
}

func TestFindings_DiscountNotSubtracted(t *testing.T) {
	db := newTestDB(t)
	discountOrder(t, db, 190, 5)
	perLine := discountOrder(t, db, 195, 5)
	plain := discountOrder(t, db, 200, 5)
	unknown := discountOrder(t, db, 150, 5)
	discountOrder(t, db, 200, 0) // no discounted line: never examined
	all := runFindings(t, db)
	wantFinding(t, all, KindDiscountPerLine, 1, perLine)
	wantFinding(t, all, KindDiscountNotSubtracted, 1, plain)
	wantFinding(t, all, KindDiscountUnknown, 1, unknown)
}

func TestFindings_DiscountKindsInFixedOrder(t *testing.T) {
	db := newTestDB(t)
	discountOrder(t, db, 150, 5)
	discountOrder(t, db, 200, 5)
	discountOrder(t, db, 195, 5)
	got := runFindings(t, db)
	want := []string{KindDiscountPerLine, KindDiscountNotSubtracted, KindDiscountUnknown}
	if len(got) != len(want) {
		t.Fatalf("findings = %+v, want kinds %v", got, want)
	}
	for i, kind := range want {
		if got[i].Kind != kind {
			t.Errorf("finding %d = %s, want %s", i, got[i].Kind, kind)
		}
	}
}

// bulkDiscountOrders inserts n Orders of two lines each, in batches. Every odd
// Order has a discounted line; its total cycles through the per-unit (fine),
// per-line, Σ price and unknown rules.
func bulkDiscountOrders(t *testing.T, db *mongo.Database, n int) {
	t.Helper()
	totals := []float64{190, 195, 200, 150}
	orders, lines := make([]any, 0, bulkBatch), make([]any, 0, 2*bulkBatch)
	flush := func() {
		if len(orders) == 0 {
			return
		}
		if _, err := db.Collection("orders").InsertMany(context.Background(), orders); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Collection("order_items").InsertMany(context.Background(), lines); err != nil {
			t.Fatal(err)
		}
		orders, lines = orders[:0], lines[:0]
	}
	for i := 0; i < n; i++ {
		id, total, discount := primitive.NewObjectID(), 200.0, 0.0
		if i%2 == 1 {
			total, discount = totals[(i/2)%len(totals)], 5.0
		}
		orders = append(orders, bson.M{"_id": id, "total": total})
		lines = append(lines,
			bson.M{"orderId": id, "price": 100.0, "quantity": 2, "discount": discount},
			bson.M{"orderId": id, "price": 100.0, "quantity": 2, "discount": 0.0})
		if len(orders) == bulkBatch {
			flush()
		}
	}
	flush()
}

func TestFindings_DiscountCheckIsFastWithoutAnIndex(t *testing.T) {
	db := newTestDB(t)
	bulkDiscountOrders(t, db, bulkOrders)
	ctx, cancel := context.WithTimeout(context.Background(), discountTimeLimit)
	defer cancel()

	start := time.Now()
	got, err := checkDiscountConvention(ctx, db)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("discount check after %s: %v", elapsed, err)
	}
	t.Logf("discount check on %d orders took %s", bulkOrders, elapsed)

	perKind := int64(bulkOrders / 2 / 4)
	for _, kind := range []string{KindDiscountPerLine, KindDiscountNotSubtracted, KindDiscountUnknown} {
		f, ok := findKind(got, kind)
		if !ok || f.Total != perKind || len(f.Ids) != maxFindingIds {
			t.Errorf("%s = %+v (found %v), want total %d", kind, f, ok, perKind)
		}
	}
}
