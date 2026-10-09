package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/data/ledger"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// saleFixture is a branch selling one Unit of one Product against a MongoDB
// replica set (MONGO_TEST_URI); the test is skipped without one.
type saleFixture struct {
	t        *testing.T
	pos      *mongo.Database
	orders   *orderEntity
	branchId primitive.ObjectID
	product  primitive.ObjectID
	unit     primitive.ObjectID
}

func newSaleFixture(t *testing.T) *saleFixture {
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
	pos := client.Database(fmt.Sprintf("pos_sale_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	f := &saleFixture{
		t:        t,
		pos:      pos,
		orders:   NewOrderEntity(&db.Resource{Client: client, PosDb: pos}).(*orderEntity),
		branchId: primitive.NewObjectID(),
		product:  primitive.NewObjectID(),
		unit:     primitive.NewObjectID(),
	}
	f.insert("products", bson.M{"_id": f.product, "name": "Paracetamol", "soldFirst": 0})
	f.insert("product_units", entities.ProductUnit{Id: f.unit, ProductId: f.product, Unit: "TAB", CostPrice: 3})
	f.insert("product_prices", entities.ProductPrice{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: f.unit, CustomerType: "General", Price: 10})
	f.insert("product_prices", entities.ProductPrice{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: f.unit, CustomerType: "Wholesaler", Price: 8})
	return f
}

// sell records a Sale through the Stock ledger, as the order handler does.
func (f *saleFixture) sell(form request.Sale) (*ledger.Sold, error) {
	return newLedger(&db.Resource{Client: f.pos.Client(), PosDb: f.pos}).Sell(context.Background(), form)
}

func (f *saleFixture) insert(collection string, doc any) {
	f.t.Helper()
	if _, err := f.pos.Collection(collection).InsertOne(context.Background(), doc); err != nil {
		f.t.Fatal(err)
	}
}

func (f *saleFixture) stock(seq, qty int, cost float64) primitive.ObjectID {
	id := primitive.NewObjectID()
	f.insert("product_stocks", entities.ProductStock{Id: id, BranchId: f.branchId, ProductId: f.product, UnitId: f.unit, Sequence: seq, Quantity: qty, CostPrice: cost})
	return id
}

func (f *saleFixture) quantity(stock primitive.ObjectID) int {
	f.t.Helper()
	var s entities.ProductStock
	if err := f.pos.Collection("product_stocks").FindOne(context.Background(), bson.M{"_id": stock}).Decode(&s); err != nil {
		f.t.Fatal(err)
	}
	return s.Quantity
}

func (f *saleFixture) soldFirst() int {
	f.t.Helper()
	var p struct {
		SoldFirst int `bson:"soldFirst"`
	}
	if err := f.pos.Collection("products").FindOne(context.Background(), bson.M{"_id": f.product}).Decode(&p); err != nil {
		f.t.Fatal(err)
	}
	return p.SoldFirst
}

func (f *saleFixture) count(collection string) int64 {
	f.t.Helper()
	n, err := f.pos.Collection(collection).CountDocuments(context.Background(), bson.M{})
	if err != nil {
		f.t.Fatal(err)
	}
	return n
}

func (f *saleFixture) items(order primitive.ObjectID) []entities.OrderItem {
	f.t.Helper()
	cursor, err := f.pos.Collection("order_items").Find(context.Background(), bson.M{"orderId": order}, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		f.t.Fatal(err)
	}
	var items []entities.OrderItem
	if err := cursor.All(context.Background(), &items); err != nil {
		f.t.Fatal(err)
	}
	return items
}

func (f *saleFixture) sale(id string, paid float64, lines ...request.SaleLine) request.Sale {
	for i := range lines {
		lines[i].ProductId, lines[i].UnitId = f.product.Hex(), f.unit.Hex()
	}
	return request.Sale{
		SaleId: id, Items: lines, Type: "CASH",
		Payments:  []request.OrderPayment{{Amount: paid, Type: "CASH"}},
		BranchId:  f.branchId.Hex(),
		CreatedBy: "cashier",
	}
}

func TestRecordSalePricesAndDrawsStockOnTheServer(t *testing.T) {
	f := newSaleFixture(t)
	later := f.stock(2, 3, 5)
	first := f.stock(1, 2, 4)

	got, err := f.sell(f.sale("s1", 100, request.SaleLine{Quantity: 7, PriceType: "General", Discount: 1}))
	if err != nil {
		t.Fatal(err)
	}
	o := got.Order
	// 7 × 10 less 1 a unit; cost 2×4 + 3×5 + 2×3 (Sold first at the Unit's cost).
	if o.Total != 63 || o.TotalCost != 29 || o.Discount != 7 || o.SaleId != "s1" {
		t.Fatalf("order money %+v", o)
	}
	if len(o.Payments) != 1 || o.Payments[0].Change != 37 || o.Payments[0].Total != 63 {
		t.Fatalf("payments %+v", o.Payments)
	}
	if f.quantity(first) != 0 || f.quantity(later) != 0 || f.soldFirst() != -2 {
		t.Fatalf("stock %d %d sold first %d", f.quantity(first), f.quantity(later), f.soldFirst())
	}
	items := f.items(o.Id)
	want := []entities.OrderItemStock{{StockId: first.Hex(), Quantity: 2}, {StockId: later.Hex(), Quantity: 3}, {Quantity: 2}}
	if len(items) != 1 || fmt.Sprint(items[0].Stocks) != fmt.Sprint(want) || items[0].Price != 70 || items[0].Discount != 1 || items[0].CostPrice != 29 {
		t.Fatalf("items %+v", items)
	}
	if len(got.Stocks) != 2 || f.count("product_histories") != 1 {
		t.Fatalf("stocks %+v histories %d", got.Stocks, f.count("product_histories"))
	}
}

func TestRecordSaleOversellOwesTheLastStock(t *testing.T) {
	f := newSaleFixture(t)
	lot := f.stock(1, 2, 4)

	got, err := f.sell(f.sale("s1", 50, request.SaleLine{Quantity: 5, PriceType: "General", AllowOversell: true}))
	if err != nil {
		t.Fatal(err)
	}
	items := f.items(got.Order.Id)
	if items[0].OversoldQty != 3 || fmt.Sprint(items[0].Stocks) != fmt.Sprint([]entities.OrderItemStock{{StockId: lot.Hex(), Quantity: 2}}) {
		t.Fatalf("item %+v", items[0])
	}
	if f.quantity(lot) != 0 || f.soldFirst() != 0 {
		t.Fatalf("stock %d sold first %d", f.quantity(lot), f.soldFirst())
	}
}

func TestRecordSaleLinesOfOneUnitDrawInTurn(t *testing.T) {
	f := newSaleFixture(t)
	lot := f.stock(1, 3, 4)

	got, err := f.sell(f.sale("s1", 100,
		request.SaleLine{Quantity: 2, PriceType: "General"},
		request.SaleLine{Quantity: 2, PriceType: "Wholesaler"}))
	if err != nil {
		t.Fatal(err)
	}
	items := f.items(got.Order.Id)
	if fmt.Sprint(items[1].Stocks) != fmt.Sprint([]entities.OrderItemStock{{StockId: lot.Hex(), Quantity: 1}, {Quantity: 1}}) {
		t.Fatalf("second line %+v", items[1].Stocks)
	}
	if got.Order.Total != 36 || f.quantity(lot) != 0 || f.soldFirst() != -1 {
		t.Fatalf("total %v stock %d sold first %d", got.Order.Total, f.quantity(lot), f.soldFirst())
	}
}

func TestRecordSaleRepeatReturnsTheRecordedOrder(t *testing.T) {
	f := newSaleFixture(t)
	lot := f.stock(1, 10, 4)
	s := f.sale("s1", 50, request.SaleLine{Quantity: 2, PriceType: "General"})

	first, err := f.sell(s)
	if err != nil {
		t.Fatal(err)
	}
	again, err := f.sell(s)
	if err != nil {
		t.Fatal(err)
	}
	if again.Order.Id != first.Order.Id || len(again.Order.Payments) != 1 || len(again.Stocks) != 1 {
		t.Fatalf("repeat recorded %+v", again)
	}
	if f.quantity(lot) != 8 || f.count("orders") != 1 {
		t.Fatalf("stock %d orders %d", f.quantity(lot), f.count("orders"))
	}

	other := f.sale("s1", 50, request.SaleLine{Quantity: 3, PriceType: "General"})
	if _, err := f.sell(other); !errors.Is(err, ledger.ErrSaleConflict) {
		t.Fatalf("expected ErrSaleConflict, got %v", err)
	}
}

func TestRecordSaleConcurrentRepeatsRecordOneOrder(t *testing.T) {
	f := newSaleFixture(t)
	lot := f.stock(1, 10, 4)
	s := f.sale("s1", 50, request.SaleLine{Quantity: 2, PriceType: "General"})

	var wg sync.WaitGroup
	ids := make([]primitive.ObjectID, 4)
	errs := make([]error, 4)
	for i := range ids {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			r, err := f.sell(s)
			errs[i] = err
			if r != nil {
				ids[i] = r.Order.Id
			}
		}(i)
	}
	wg.Wait()
	for i := range ids {
		if errs[i] != nil || ids[i] != ids[0] {
			t.Fatalf("attempt %d: %v %v (first %v)", i, ids[i], errs[i], ids[0])
		}
	}
	if f.quantity(lot) != 8 || f.count("orders") != 1 {
		t.Fatalf("stock %d orders %d", f.quantity(lot), f.count("orders"))
	}
}

func TestRecordSaleRejectedWritesNothing(t *testing.T) {
	f := newSaleFixture(t)
	lot := f.stock(1, 10, 4)

	underpaid := f.sale("s1", 19.99, request.SaleLine{Quantity: 2, PriceType: "General"})
	wrongUnit := f.sale("s2", 50, request.SaleLine{Quantity: 1, PriceType: "General"})
	wrongUnit.Items[0].UnitId = primitive.NewObjectID().Hex()
	secondLineBad := f.sale("s3", 50, request.SaleLine{Quantity: 1, PriceType: "General"}, request.SaleLine{Quantity: 1})
	secondLineBad.Items[1].ProductId = primitive.NewObjectID().Hex()

	for name, s := range map[string]request.Sale{"underpaid": underpaid, "wrong unit": wrongUnit, "second line bad": secondLineBad} {
		var rejected *ledger.Rejected
		if _, err := f.sell(s); !errors.As(err, &rejected) {
			t.Errorf("%s: expected SaleRejected, got %v", name, err)
		}
	}
	if f.quantity(lot) != 10 || f.count("orders") != 0 || f.count("order_items") != 0 || f.count("payments") != 0 || f.count("product_histories") != 0 {
		t.Fatalf("rejected sale wrote: stock %d orders %d", f.quantity(lot), f.count("orders"))
	}
}

func TestCancelLineRecomputesOrderMoney(t *testing.T) {
	f := newSaleFixture(t)
	f.stock(1, 10, 4)
	got, err := f.sell(f.sale("s1", 100,
		request.SaleLine{Quantity: 2, PriceType: "General", Discount: 1},
		request.SaleLine{Quantity: 3, PriceType: "Wholesaler", Discount: 0.5}))
	if err != nil {
		t.Fatal(err)
	}
	items := f.items(got.Order.Id)
	if _, err := f.orders.CancelOrderItemById(items[0].Id.Hex(), "manager", f.branchId.Hex(), "wrong item"); err != nil {
		t.Fatal(err)
	}
	var o entities.Order
	if err := f.pos.Collection("orders").FindOne(context.Background(), bson.M{"_id": got.Order.Id}).Decode(&o); err != nil {
		t.Fatal(err)
	}
	// Left: 3 × 8 less 0.5 a unit, cost 3 × 4.
	if o.Total != 22.5 || o.TotalCost != 12 || o.Discount != 1.5 {
		t.Fatalf("order after cancel %+v", o)
	}
}

func TestRecordSaleDrawsOnlyTheBranchStock(t *testing.T) {
	f := newSaleFixture(t)
	elsewhere := primitive.NewObjectID()
	f.insert("product_stocks", entities.ProductStock{Id: elsewhere, BranchId: primitive.NewObjectID(), ProductId: f.product, UnitId: f.unit, Sequence: 0, Quantity: 10, CostPrice: 4})

	if _, err := f.sell(f.sale("s1", 50, request.SaleLine{Quantity: 2, PriceType: "General"})); err != nil {
		t.Fatal(err)
	}
	if f.quantity(elsewhere) != 10 || f.soldFirst() != -2 {
		t.Fatalf("other branch stock %d sold first %d", f.quantity(elsewhere), f.soldFirst())
	}
}

func TestRecordSaleRejectsAUnitOfAnotherProduct(t *testing.T) {
	f := newSaleFixture(t)
	other := primitive.NewObjectID()
	f.insert("products", bson.M{"_id": other, "name": "Other", "soldFirst": 0})
	s := f.sale("s1", 50, request.SaleLine{Quantity: 1, PriceType: "General"})
	s.Items[0].ProductId = other.Hex()

	var rejected *ledger.Rejected
	if _, err := f.sell(s); !errors.As(err, &rejected) {
		t.Fatalf("expected SaleRejected, got %v", err)
	}
}

func TestOrdersSaleIdIsUniqueWhenPresent(t *testing.T) {
	f := newSaleFixture(t)
	orders := f.pos.Collection("orders")
	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := orders.InsertOne(ctx, bson.M{"_id": primitive.NewObjectID(), "code": "old"}); err != nil {
			t.Fatalf("orders without saleId: %v", err)
		}
	}
	if _, err := orders.InsertOne(ctx, bson.M{"_id": primitive.NewObjectID(), "saleId": "s1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := orders.InsertOne(ctx, bson.M{"_id": primitive.NewObjectID(), "saleId": "s1"}); !mongo.IsDuplicateKeyError(err) {
		t.Fatalf("expected duplicate key, got %v", err)
	}
}

func TestRecordSaleRefusedTakesNoOrderCode(t *testing.T) {
	f := newSaleFixture(t)
	f.stock(1, 10, 4)

	var rejected *ledger.Rejected
	if _, err := f.sell(f.sale("refused", 1, request.SaleLine{Quantity: 2, PriceType: "General"})); !errors.As(err, &rejected) {
		t.Fatalf("expected an underpaid Sale to be refused, got %v", err)
	}
	got, err := f.sell(f.sale("accepted", 50, request.SaleLine{Quantity: 2, PriceType: "General"}))
	if err != nil {
		t.Fatal(err)
	}
	var seq entities.Sequence
	if err := f.pos.Collection("sequences").FindOne(context.Background(), bson.M{"field": constant.ORDER}).Decode(&seq); err != nil {
		t.Fatal(err)
	}
	if seq.Value != 1 || got.Order.Code != seq.GenerateCode() {
		t.Fatalf("Order code %q, sequence at %d: the refused Sale must not have taken a code", got.Order.Code, seq.Value)
	}
}

func TestRepairOrderTotalsAgreesWithACancelledLine(t *testing.T) {
	f := newSaleFixture(t)
	f.stock(1, 10, 4)
	got, err := f.sell(f.sale("s1", 100,
		request.SaleLine{Quantity: 1, PriceType: "General", Discount: 3.333},
		request.SaleLine{Quantity: 1, PriceType: "General", Discount: 3.333},
		request.SaleLine{Quantity: 1, PriceType: "General", Discount: 3.333}))
	if err != nil {
		t.Fatal(err)
	}
	items := f.items(got.Order.Id)
	if _, err := f.orders.CancelOrderItemById(items[0].Id.Hex(), "manager", f.branchId.Hex(), ""); err != nil {
		t.Fatal(err)
	}
	report, err := RepairCancelledLineTotals(context.Background(), f.pos, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(report) != 0 {
		t.Fatalf("repair-order-totals flags an Order the cancel just recomputed: %+v", report)
	}
}

func TestCancelByOrderAndProductCancelsTheLineStillStanding(t *testing.T) {
	f := newSaleFixture(t)
	f.stock(1, 10, 4)
	got, err := f.sell(f.sale("s1", 100,
		request.SaleLine{Quantity: 1, PriceType: "General"},
		request.SaleLine{Quantity: 2, PriceType: "Wholesaler"}))
	if err != nil {
		t.Fatal(err)
	}
	items := f.items(got.Order.Id)
	if _, err := f.orders.CancelOrderItemById(items[0].Id.Hex(), "manager", f.branchId.Hex(), ""); err != nil {
		t.Fatal(err)
	}
	cancelled, err := f.orders.CancelOrderItemByOrderProductId(got.Order.Id.Hex(), f.product.Hex(), "manager", f.branchId.Hex(), "")
	if err != nil {
		t.Fatalf("the second Line of the Product must still be cancellable, got %v", err)
	}
	if cancelled.Id != items[1].Id || cancelled.Status != constant.CANCELLED {
		t.Fatalf("cancelled %+v, want the second Line", cancelled)
	}
}
