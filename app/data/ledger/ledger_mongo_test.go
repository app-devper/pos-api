package ledger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// The ledger is tested through its methods against a real replica set
// (MONGO_TEST_URI): transactions are part of what it promises (ADR-0001).
type fixture struct {
	t       *testing.T
	pos     *mongo.Database
	ledger  *Ledger
	branch  primitive.ObjectID
	product primitive.ObjectID
	tab     primitive.ObjectID // the Product's main Unit
	box     primitive.ObjectID
}

func newFixture(t *testing.T) *fixture {
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
	pos := client.Database(fmt.Sprintf("pos_ledger_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		_ = pos.Drop(ctx)
		_ = client.Disconnect(ctx)
	})
	var seq atomic.Int64
	codes := func(_ context.Context, _ string, prefix string) (string, error) {
		return fmt.Sprintf("%s%04d", prefix, seq.Add(1)), nil
	}
	f := &fixture{t: t, pos: pos, ledger: New(client, pos, codes), branch: primitive.NewObjectID(),
		product: primitive.NewObjectID(), tab: primitive.NewObjectID(), box: primitive.NewObjectID()}
	f.insert("products", entities.Product{Id: f.product, Name: "Paracetamol", Unit: "TAB"})
	f.insert("product_units", entities.ProductUnit{Id: f.tab, ProductId: f.product, Unit: "TAB", CostPrice: 3})
	f.insert("product_units", entities.ProductUnit{Id: f.box, ProductId: f.product, Unit: "BOX", CostPrice: 30})
	return f
}

func (f *fixture) insert(collection string, doc any) {
	f.t.Helper()
	if _, err := f.pos.Collection(collection).InsertOne(context.Background(), doc); err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) stock(unit primitive.ObjectID, seq, qty int) primitive.ObjectID {
	id := primitive.NewObjectID()
	f.insert("product_stocks", entities.ProductStock{Id: id, BranchId: f.branch, ProductId: f.product, UnitId: unit, Sequence: seq, Quantity: qty, CostPrice: 3})
	return id
}

// owed inserts a confirmed Line in a branch that is still owed qty of a Unit.
func (f *fixture) owed(branch, unit primitive.ObjectID, qty int) primitive.ObjectID {
	id := primitive.NewObjectID()
	f.insert("order_items", entities.OrderItem{Id: id, OrderId: primitive.NewObjectID(), BranchId: branch, ProductId: f.product, UnitId: unit, Quantity: qty, Status: constant.CONFIRMED, OversoldQty: qty})
	return id
}

func (f *fixture) quantity(stock primitive.ObjectID) int {
	f.t.Helper()
	var got entities.ProductStock
	if err := f.pos.Collection("product_stocks").FindOne(context.Background(), bson.M{"_id": stock}).Decode(&got); err != nil {
		f.t.Fatal(err)
	}
	return got.Quantity
}

func (f *fixture) line(id primitive.ObjectID) entities.OrderItem {
	f.t.Helper()
	var got entities.OrderItem
	if err := f.pos.Collection("order_items").FindOne(context.Background(), bson.M{"_id": id}).Decode(&got); err != nil {
		f.t.Fatal(err)
	}
	return got
}

func (f *fixture) count(collection string, filter bson.M) int64 {
	f.t.Helper()
	n, err := f.pos.Collection(collection).CountDocuments(context.Background(), filter)
	if err != nil {
		f.t.Fatal(err)
	}
	return n
}

func (f *fixture) histories() []entities.ProductHistory {
	f.t.Helper()
	cursor, err := f.pos.Collection("product_histories").Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		f.t.Fatal(err)
	}
	var got []entities.ProductHistory
	if err := cursor.All(context.Background(), &got); err != nil {
		f.t.Fatal(err)
	}
	return got
}

// failInserts makes every later insert into a collection fail, so a command
// that has already moved Stock must roll back.
func (f *fixture) failInserts(collection string) {
	f.t.Helper()
	_ = f.pos.CreateCollection(context.Background(), collection)
	err := f.pos.RunCommand(context.Background(), bson.D{{Key: "collMod", Value: collection}, {Key: "validator", Value: bson.M{"_id": bson.M{"$exists": false}}}, {Key: "validationLevel", Value: "strict"}}).Err()
	if err != nil {
		f.t.Fatal(err)
	}
}

func (f *fixture) adjust(stock primitive.ObjectID, delta int) request.StockAdjustment {
	return request.StockAdjustment{BranchId: f.branch.Hex(), ProductId: f.product.Hex(), StockId: stock.Hex(), Delta: delta, Reason: constant.AdjustmentReasonCount, CreatedBy: "admin"}
}

func (f *fixture) receive(status string, branch primitive.ObjectID, quantities ...int) primitive.ObjectID {
	id := primitive.NewObjectID()
	f.insert("receives", entities.Receive{Id: id, BranchId: branch, Code: "RC-1", Status: status})
	for i, qty := range quantities {
		f.insert("receive_items", entities.ReceiveItem{ReceiveId: id, ProductId: f.product, Quantity: qty, CostPrice: 4, LotNumber: fmt.Sprintf("L%d", i+1), ExpireDate: time.Now().AddDate(1, 0, 0)})
	}
	return id
}

func TestAdjustUpSettlesOversellOfTheSameUnitDrawingFromTheStock(t *testing.T) {
	f := newFixture(t)
	stock := f.stock(f.tab, 1, 2)
	first, second := f.owed(f.branch, f.tab, 2), f.owed(f.branch, f.tab, 2)
	boxLine := f.owed(f.branch, f.box, 5)
	otherBranch := f.owed(primitive.NewObjectID(), f.tab, 4)

	got, err := f.ledger.Adjust(context.Background(), f.adjust(stock, 3))
	if err != nil {
		t.Fatal(err)
	}
	if got.Before != 2 || got.After != 5 {
		t.Fatalf("Adjustment recorded %d → %d, want 2 → 5", got.Before, got.After)
	}
	if q := f.quantity(stock); q != 2 {
		t.Fatalf("Stock = %d, want 2: the 3 adjusted in go to the waiting Lines", q)
	}
	a, b := f.line(first), f.line(second)
	if a.OversoldQty != 0 || len(a.Stocks) != 1 || a.Stocks[0] != (entities.OrderItemStock{StockId: stock.Hex(), Quantity: 2}) {
		t.Fatalf("oldest Line not settled from the Stock: %+v", a)
	}
	if b.OversoldQty != 1 || len(b.Stocks) != 1 || b.Stocks[0].Quantity != 1 {
		t.Fatalf("second Line not partly settled: %+v", b)
	}
	if f.line(boxLine).OversoldQty != 5 {
		t.Fatal("a TAB Stock settled a BOX Line")
	}
	if f.line(otherBranch).OversoldQty != 4 {
		t.Fatal("another branch's Line was settled")
	}
	h := f.histories()
	if len(h) != 1 || h[0].Balance != 2 || h[0].Unit != "TAB" || !strings.Contains(h[0].Description, "3") {
		t.Fatalf("want one history row with the final balance, got %+v", h)
	}
}

func TestAdjustDownNeverSettlesAndNeverGoesBelowZero(t *testing.T) {
	f := newFixture(t)
	stock := f.stock(f.tab, 1, 5)
	waiting := f.owed(f.branch, f.tab, 2)

	if _, err := f.ledger.Adjust(context.Background(), f.adjust(stock, -2)); err != nil {
		t.Fatal(err)
	}
	if f.quantity(stock) != 3 || f.line(waiting).OversoldQty != 2 {
		t.Fatal("a lowering Adjustment settled a debt")
	}
	_, err := f.ledger.Adjust(context.Background(), f.adjust(stock, -4))
	var rejected *Rejected
	if !errors.As(err, &rejected) {
		t.Fatalf("want Rejected for going below zero, got %v", err)
	}
	if f.quantity(stock) != 3 || f.count("stock_adjustments", bson.M{}) != 1 {
		t.Fatal("rejected Adjustment left effects")
	}
	for _, bad := range []request.StockAdjustment{
		func() request.StockAdjustment { r := f.adjust(stock, 0); return r }(),
		func() request.StockAdjustment { r := f.adjust(stock, 1); r.Reason = "made up"; return r }(),
	} {
		if _, err := f.ledger.Adjust(context.Background(), bad); !errors.As(err, &rejected) {
			t.Fatalf("want Rejected, got %v", err)
		}
	}
	other := f.adjust(stock, 1)
	other.BranchId = primitive.NewObjectID().Hex()
	if _, err := f.ledger.Adjust(context.Background(), other); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another branch's Stock: want ErrNotFound, got %v", err)
	}
}

func TestCountThatRaisesAStockSettlesTheSameUnit(t *testing.T) {
	f := newFixture(t)
	raised, lowered := f.stock(f.tab, 1, 1), f.stock(f.tab, 2, 6)
	waiting := f.owed(f.branch, f.tab, 3)

	got, err := f.ledger.Count(context.Background(), request.StockCount{BranchId: f.branch.Hex(), CreatedBy: "admin", Items: []request.StockCountItem{
		{ProductId: f.product.Hex(), StockId: raised.Hex(), Counted: 5},
		{ProductId: f.product.Hex(), StockId: lowered.Hex(), Counted: 4},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Items) != 2 || got.Items[0].Delta != 4 || got.Items[1].Delta != -2 {
		t.Fatalf("Count Lines incorrect: %+v", got.Items)
	}
	if f.quantity(raised) != 2 || f.quantity(lowered) != 4 || f.line(waiting).OversoldQty != 0 {
		t.Fatalf("raised=%d lowered=%d owed=%d, want 2, 4, 0", f.quantity(raised), f.quantity(lowered), f.line(waiting).OversoldQty)
	}
	if f.count("stock_adjustments", bson.M{}) != 2 || len(f.histories()) != 2 {
		t.Fatal("want one Adjustment and one history row per changed Stock")
	}
}

func TestImportReceiveSettlesInItsOwnTransactionAndMarksItImported(t *testing.T) {
	f := newFixture(t)
	waiting := f.owed(f.branch, f.tab, 3)
	elsewhere := f.owed(primitive.NewObjectID(), f.tab, 1)
	f.stock(f.tab, 4, 0)
	id := f.receive(constant.ACTIVE, f.branch, 10, 5)

	got, err := f.ledger.ImportReceive(context.Background(), id.Hex(), f.branch.Hex(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != constant.IMPORTED || got.TotalCost != 60 || len(got.Items) != 2 {
		t.Fatalf("Receive not marked imported: %+v", got)
	}
	cursor, err := f.pos.Collection("product_stocks").Find(context.Background(), bson.M{"receiveCode": "RC-1"}, options.Find().SetSort(bson.D{{Key: "sequence", Value: 1}}))
	if err != nil {
		t.Fatal(err)
	}
	var stocks []entities.ProductStock
	if err := cursor.All(context.Background(), &stocks); err != nil {
		t.Fatal(err)
	}
	if len(stocks) != 2 || stocks[0].Quantity != 7 || stocks[0].Import != 10 || stocks[1].Quantity != 5 || stocks[0].UnitId != f.tab {
		t.Fatalf("new Stocks incorrect: %+v", stocks)
	}
	if stocks[0].Sequence != 5 || stocks[1].Sequence != 6 {
		t.Fatalf("new Stocks must follow the branch's existing sequence, got %d and %d", stocks[0].Sequence, stocks[1].Sequence)
	}
	settled := f.line(waiting)
	if settled.OversoldQty != 0 || len(settled.Stocks) != 1 || settled.Stocks[0].StockId != stocks[0].Id.Hex() {
		t.Fatalf("waiting Line not settled from the first new Stock: %+v", settled)
	}
	if f.line(elsewhere).OversoldQty != 1 {
		t.Fatal("a Receive settled another branch's Line")
	}
	h := f.histories()
	if len(h) != 2 || h[0].Import != 10 || h[0].Balance != 12 || h[1].Balance != 12 || !strings.Contains(h[0].Description, "ส่งของค้างลูกค้า 3") {
		t.Fatalf("want one history row per new Stock, each with the balance once the import is done, got %+v", h)
	}
}

func TestImportReceiveRefusesImportedCancelledAndOtherBranchReceives(t *testing.T) {
	f := newFixture(t)
	var rejected *Rejected
	for _, status := range []string{constant.IMPORTED, constant.CANCELLED} {
		id := f.receive(status, f.branch, 3)
		if _, err := f.ledger.ImportReceive(context.Background(), id.Hex(), f.branch.Hex(), "admin"); !errors.As(err, &rejected) {
			t.Fatalf("%s Receive: want Rejected, got %v", status, err)
		}
	}
	id := f.receive(constant.ACTIVE, primitive.NewObjectID(), 3)
	if _, err := f.ledger.ImportReceive(context.Background(), id.Hex(), f.branch.Hex(), "admin"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another branch's Receive: want ErrNotFound, got %v", err)
	}
	if f.count("product_stocks", bson.M{}) != 0 || len(f.histories()) != 0 {
		t.Fatal("a refused import left Stock or history")
	}
}

func TestImportReceiveRollsBackEverythingWhenAWriteFails(t *testing.T) {
	f := newFixture(t)
	waiting := f.owed(f.branch, f.tab, 2)
	id := f.receive(constant.ACTIVE, f.branch, 5)
	f.failInserts("product_histories")

	if _, err := f.ledger.ImportReceive(context.Background(), id.Hex(), f.branch.Hex(), "admin"); err == nil {
		t.Fatal("expected the failed history write to fail the import")
	}
	var receive entities.Receive
	if err := f.pos.Collection("receives").FindOne(context.Background(), bson.M{"_id": id}).Decode(&receive); err != nil {
		t.Fatal(err)
	}
	if receive.Status != constant.ACTIVE || f.count("product_stocks", bson.M{}) != 0 || f.line(waiting).OversoldQty != 2 {
		t.Fatal("failed import left partial effects")
	}
}

func TestReturnPutsGoodsBackThenServesOtherWaitingLines(t *testing.T) {
	f := newFixture(t)
	stock := f.stock(f.tab, 1, 0)
	order, sold := primitive.NewObjectID(), primitive.NewObjectID()
	f.insert("orders", entities.Order{Id: order, BranchId: f.branch, Status: constant.CONFIRMED})
	f.insert("order_items", entities.OrderItem{Id: sold, OrderId: order, BranchId: f.branch, ProductId: f.product, UnitId: f.tab, Quantity: 3, Price: 30, Status: constant.CONFIRMED, Stocks: []entities.OrderItemStock{{StockId: stock.Hex(), Quantity: 3}}})
	waiting := f.owed(f.branch, f.tab, 1)

	got, err := f.ledger.Return(context.Background(), request.ProductReturn{OrderId: order.Hex(), BranchId: f.branch.Hex(), CreatedBy: "admin", Items: []request.ProductReturnItem{{OrderItemId: sold.Hex(), Quantity: 2, Refund: 20}}})
	if err != nil {
		t.Fatal(err)
	}
	if got.TotalRefund != 20 || f.line(sold).ReturnedQty != 2 {
		t.Fatalf("Return not recorded: %+v", got)
	}
	if f.quantity(stock) != 1 || f.line(waiting).OversoldQty != 0 {
		t.Fatalf("Stock=%d owed=%d, want 1 and 0: returned goods serve the waiting Line", f.quantity(stock), f.line(waiting).OversoldQty)
	}
	if h := f.histories(); len(h) != 1 || h[0].Balance != 1 {
		t.Fatalf("want one history row with the final balance, got %+v", h)
	}
	_, err = f.ledger.Return(context.Background(), request.ProductReturn{OrderId: order.Hex(), BranchId: f.branch.Hex(), Items: []request.ProductReturnItem{{OrderItemId: sold.Hex(), Quantity: 2}}})
	var rejected *Rejected
	if !errors.As(err, &rejected) || !strings.Contains(rejected.Reason, "คืนได้สูงสุด 1") {
		t.Fatalf("want the Thai cap refusal, got %v", err)
	}
}

func TestRepairCrossUnitOversellReportsThenPutsTheDrawBack(t *testing.T) {
	f := newFixture(t)
	tabStock := f.stock(f.tab, 1, 8)
	boxLine := primitive.NewObjectID()
	f.insert("order_items", entities.OrderItem{Id: boxLine, OrderId: primitive.NewObjectID(), BranchId: f.branch, ProductId: f.product, UnitId: f.box, Quantity: 2, Status: constant.CONFIRMED,
		Stocks: []entities.OrderItemStock{{StockId: tabStock.Hex(), Quantity: 2}}})
	sameUnit := f.owed(f.branch, f.tab, 0)
	f.pos.Collection("order_items").UpdateOne(context.Background(), bson.M{"_id": sameUnit}, bson.M{"$set": bson.M{"stocks": []entities.OrderItemStock{{StockId: tabStock.Hex(), Quantity: 1}}}})

	report, err := f.ledger.RepairCrossUnitOversell(context.Background(), false, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if len(report) != 1 || report[0].Line != boxLine || report[0].Stock != tabStock || report[0].Quantity != 2 {
		t.Fatalf("report = %+v, want the one BOX Line drawn from a TAB Stock", report)
	}
	if f.quantity(tabStock) != 8 || f.line(boxLine).OversoldQty != 0 {
		t.Fatal("a report-only run wrote something")
	}

	if _, err := f.ledger.RepairCrossUnitOversell(context.Background(), true, "admin"); err != nil {
		t.Fatal(err)
	}
	fixed := f.line(boxLine)
	if f.quantity(tabStock) != 10 || fixed.OversoldQty != 2 || len(fixed.Stocks) != 0 {
		t.Fatalf("Stock=%d Line=%+v, want the 2 TAB back and the BOX Line owed 2 again", f.quantity(tabStock), fixed)
	}
	if len(f.line(sameUnit).Stocks) != 1 {
		t.Fatal("a same-Unit settlement was undone")
	}
	if again, _ := f.ledger.RepairCrossUnitOversell(context.Background(), false, "admin"); len(again) != 0 {
		t.Fatal("repair is not idempotent")
	}
}

func TestReturnNeverSettlesTheReturnedLineItself(t *testing.T) {
	f := newFixture(t)
	stock := f.stock(f.tab, 1, 0)
	order, line := primitive.NewObjectID(), primitive.NewObjectID()
	f.insert("orders", entities.Order{Id: order, BranchId: f.branch, Status: constant.CONFIRMED})
	// Sold 5, only 3 came from a Stock; 2 are still owed to this customer.
	f.insert("order_items", entities.OrderItem{Id: line, OrderId: order, BranchId: f.branch, ProductId: f.product, UnitId: f.tab, Quantity: 5, Price: 50, Status: constant.CONFIRMED,
		OversoldQty: 2, Stocks: []entities.OrderItemStock{{StockId: stock.Hex(), Quantity: 3}}})

	if _, err := f.ledger.Return(context.Background(), request.ProductReturn{OrderId: order.Hex(), BranchId: f.branch.Hex(), Items: []request.ProductReturnItem{{OrderItemId: line.Hex(), Quantity: 3}}}); err != nil {
		t.Fatal(err)
	}
	got := f.line(line)
	if f.quantity(stock) != 3 || got.OversoldQty != 2 || len(got.Stocks) != 1 {
		t.Fatalf("Stock=%d Line=%+v: the returned goods must not pay this Line's own debt", f.quantity(stock), got)
	}
	if _, err := f.ledger.Return(context.Background(), request.ProductReturn{OrderId: order.Hex(), BranchId: f.branch.Hex(), Items: []request.ProductReturnItem{{OrderItemId: line.Hex(), Quantity: 1}}}); err == nil {
		t.Fatal("goods the customer never received were returned again")
	}
}

func TestRepairLeavesCancelledLinesAndLinesWithoutAUnitAlone(t *testing.T) {
	f := newFixture(t)
	boxStock := f.stock(f.box, 1, 0)
	cancelled := primitive.NewObjectID()
	f.insert("order_items", entities.OrderItem{Id: cancelled, OrderId: primitive.NewObjectID(), BranchId: f.branch, ProductId: f.product, UnitId: f.tab, Quantity: 2, Status: constant.CANCELLED,
		Stocks: []entities.OrderItemStock{{StockId: boxStock.Hex(), Quantity: 2}}})
	_, err := f.pos.Collection("order_items").InsertOne(context.Background(), bson.M{"_id": primitive.NewObjectID(), "branchId": f.branch, "productId": f.product, "quantity": 1, "status": constant.CONFIRMED,
		"stocks": bson.A{bson.M{"stockid": boxStock.Hex(), "quantity": 1}}})
	if err != nil {
		t.Fatal(err)
	}

	report, err := f.ledger.RepairCrossUnitOversell(context.Background(), true, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if len(report) != 0 || f.quantity(boxStock) != 0 {
		t.Fatalf("report=%+v Stock=%d: cancelled Lines already gave their draws back, and a Line with no Unit cannot be judged", report, f.quantity(boxStock))
	}
}

func TestAReturnOfSeveralLinesNeverSettlesAnyOfThemWhateverTheOrder(t *testing.T) {
	f := newFixture(t)
	stock := f.stock(f.tab, 1, 0)
	order, a, b := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	f.insert("orders", entities.Order{Id: order, BranchId: f.branch, Status: constant.CONFIRMED})
	f.insert("order_items", entities.OrderItem{Id: a, OrderId: order, BranchId: f.branch, ProductId: f.product, UnitId: f.tab, Quantity: 3, Price: 30, Status: constant.CONFIRMED,
		Stocks: []entities.OrderItemStock{{StockId: stock.Hex(), Quantity: 3}}})
	f.insert("order_items", entities.OrderItem{Id: b, OrderId: order, BranchId: f.branch, ProductId: f.product, UnitId: f.tab, Quantity: 3, Price: 30, Status: constant.CONFIRMED,
		OversoldQty: 2, Stocks: []entities.OrderItemStock{{StockId: stock.Hex(), Quantity: 1}}})

	if _, err := f.ledger.Return(context.Background(), request.ProductReturn{OrderId: order.Hex(), BranchId: f.branch.Hex(), Items: []request.ProductReturnItem{
		{OrderItemId: a.Hex(), Quantity: 2}, {OrderItemId: b.Hex(), Quantity: 1},
	}}); err != nil {
		t.Fatal(err)
	}
	if got := f.line(b); got.OversoldQty != 2 || len(got.Stocks) != 1 || f.quantity(stock) != 3 {
		t.Fatalf("Stock=%d B=%+v: goods coming back on this Return paid a Line of the same Return", f.quantity(stock), got)
	}
}

func TestImportReceiveWithoutAStatusIsImported(t *testing.T) {
	f := newFixture(t)
	id := primitive.NewObjectID()
	if _, err := f.pos.Collection("receives").InsertOne(context.Background(), bson.M{"_id": id, "branchId": f.branch, "code": "RC-OLD"}); err != nil {
		t.Fatal(err)
	}
	f.insert("receive_items", entities.ReceiveItem{ReceiveId: id, ProductId: f.product, Quantity: 4, CostPrice: 2})
	got, err := f.ledger.ImportReceive(context.Background(), id.Hex(), f.branch.Hex(), "admin")
	if err != nil || got.Status != constant.IMPORTED {
		t.Fatalf("a Receive saved before statuses existed must import, got %+v, %v", got, err)
	}
}

// --- Sale and cancel (PR-2) ---

func (f *fixture) priced(unit primitive.ObjectID, price float64) {
	f.insert("product_prices", entities.ProductPrice{Id: primitive.NewObjectID(), ProductId: f.product, UnitId: unit, CustomerType: "General", Price: price})
}

func (f *fixture) sale(id string, paid float64, lines ...request.SaleLine) request.Sale {
	for i := range lines {
		lines[i].ProductId = f.product.Hex()
		if lines[i].UnitId == "" {
			lines[i].UnitId = f.tab.Hex()
		}
		if lines[i].PriceType == "" {
			lines[i].PriceType = "General"
		}
	}
	return request.Sale{SaleId: id, Items: lines, Type: "CASH", BranchId: f.branch.Hex(), CreatedBy: "cashier",
		Payments: []request.OrderPayment{{Amount: paid, Type: "CASH"}}}
}

func TestSellRepeatWithADifferentTenderReturnsTheRecordedOrder(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	lot := f.stock(f.tab, 1, 5)

	first, err := f.ledger.Sell(context.Background(), f.sale("s1", 50, request.SaleLine{Quantity: 2}))
	if err != nil {
		t.Fatal(err)
	}
	// The reply was lost; the till resends with the cashier's corrected tender.
	retry := f.sale("s1", 100, request.SaleLine{Quantity: 2})
	retry.Type, retry.Message = "TRANSFER", "ลูกค้าโอน"
	again, err := f.ledger.Sell(context.Background(), retry)
	if err != nil {
		t.Fatalf("a repeat that differs only in how it was paid must return the Order, got %v", err)
	}
	if again.Order.Id != first.Order.Id || f.quantity(lot) != 3 || f.count("orders", bson.M{}) != 1 {
		t.Fatal("the repeat recorded a second Order or drew Stock again")
	}
	if _, err := f.ledger.Sell(context.Background(), f.sale("s1", 50, request.SaleLine{Quantity: 3})); !errors.Is(err, ErrSaleConflict) {
		t.Fatalf("different Lines under the same id: want ErrSaleConflict, got %v", err)
	}
}

func TestSellRecognisesAnOrderRecordedWithTheOldFingerprint(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	f.stock(f.tab, 1, 5)
	s := f.sale("s-old", 50, request.SaleLine{Quantity: 1})
	old := primitive.NewObjectID()
	f.insert("orders", entities.Order{Id: old, BranchId: f.branch, Status: constant.CONFIRMED, SaleId: "s-old", SaleFingerprint: legacyFingerprint(s)})

	got, err := f.ledger.Sell(context.Background(), s)
	if err != nil {
		t.Fatal(err)
	}
	if got.Order.Id != old || f.count("orders", bson.M{}) != 1 {
		t.Fatal("a Sale recorded before the fingerprint changed was recorded again")
	}
}

func TestSellOversoldLineDrawsWhatExistsAndOwesTheRest(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	lot := f.stock(f.tab, 1, 2)

	got, err := f.ledger.Sell(context.Background(), f.sale("s1", 50, request.SaleLine{Quantity: 5, AllowOversell: true}))
	if err != nil {
		t.Fatal(err)
	}
	line := f.line(got.Lines[0].Id)
	if f.quantity(lot) != 0 || line.OversoldQty != 3 || len(line.Stocks) != 1 || line.Stocks[0].Quantity != 2 {
		t.Fatalf("Stock=%d Line=%+v", f.quantity(lot), line)
	}
	if h := f.histories(); len(h) != 1 || h[0].Quantity != 5 || h[0].Balance != 0 {
		t.Fatalf("want one history row for the Line, got %+v", h)
	}
}

func TestCancelLinePutsBackEveryDrawAndServesOtherWaitingLines(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	lot := f.stock(f.tab, 1, 1)
	sold, err := f.ledger.Sell(context.Background(), f.sale("s1", 50, request.SaleLine{Quantity: 3, AllowOversell: true}))
	if err != nil {
		t.Fatal(err)
	}
	cancelled := sold.Lines[0].Id // drew 1, owes 2
	if _, err := f.ledger.Adjust(context.Background(), f.adjust(lot, 1)); err != nil {
		t.Fatal(err) // settles 1 of the 2 owed, from lot
	}
	waiting := f.owed(f.branch, f.tab, 2)

	if err := f.ledger.CancelLine(context.Background(), cancelled.Hex(), f.branch.Hex(), "manager", "wrong item"); err != nil {
		t.Fatal(err)
	}
	// The cancelled Line drew 1 at the sale and 1 by settlement: 2 come back,
	// and both go to the Line still waiting. Its own remaining debt vanishes.
	if got := f.line(cancelled); got.Status != constant.CANCELLED {
		t.Fatalf("Line not cancelled: %+v", got)
	}
	if f.quantity(lot) != 0 || f.line(waiting).OversoldQty != 0 {
		t.Fatalf("Stock=%d waiting owed=%d, want 0 and 0", f.quantity(lot), f.line(waiting).OversoldQty)
	}
	var rejected *Rejected
	if err := f.ledger.CancelLine(context.Background(), cancelled.Hex(), f.branch.Hex(), "manager", "again"); !errors.As(err, &rejected) {
		t.Fatalf("a second cancel must be refused, got %v", err)
	}
}

func TestCancelLineRecomputesTheOrderRounded(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	f.stock(f.tab, 1, 10)
	sold, err := f.ledger.Sell(context.Background(), f.sale("s1", 100,
		request.SaleLine{Quantity: 1, Discount: 3.333},
		request.SaleLine{Quantity: 2, Discount: 3.333}))
	if err != nil {
		t.Fatal(err)
	}
	if err := f.ledger.CancelLine(context.Background(), sold.Lines[0].Id.Hex(), f.branch.Hex(), "manager", ""); err != nil {
		t.Fatal(err)
	}
	var o entities.Order
	if err := f.pos.Collection("orders").FindOne(context.Background(), bson.M{"_id": sold.Order.Id}).Decode(&o); err != nil {
		t.Fatal(err)
	}
	// What remains: 2 × 10 less 3.333 a unit = 13.334 → 13.33 (as when sold).
	if o.Total != 13.33 || o.Discount != 6.67 {
		t.Fatalf("Order after cancel total=%v discount=%v, want 13.33 and 6.67", o.Total, o.Discount)
	}
}

func TestCancelOrderGivesSoldFirstBackAndCancelsOnce(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	sold, err := f.ledger.Sell(context.Background(), f.sale("s1", 50, request.SaleLine{Quantity: 2}))
	if err != nil {
		t.Fatal(err)
	}
	var p entities.Product
	_ = f.pos.Collection("products").FindOne(context.Background(), bson.M{"_id": f.product}).Decode(&p)
	if p.SoldFirst != -2 {
		t.Fatalf("Sold first = %d after selling 2 with no Stock, want -2", p.SoldFirst)
	}
	if err := f.ledger.CancelOrder(context.Background(), sold.Order.Id.Hex(), f.branch.Hex(), "manager", "customer left"); err != nil {
		t.Fatal(err)
	}
	_ = f.pos.Collection("products").FindOne(context.Background(), bson.M{"_id": f.product}).Decode(&p)
	if p.SoldFirst != 0 || f.count("payments", bson.M{"status": constant.CANCELLED}) != 1 || f.count("order_items", bson.M{"status": constant.CANCELLED}) != 1 {
		t.Fatal("cancel did not give Sold first back or cancel the payment and Line")
	}
	var rejected *Rejected
	if err := f.ledger.CancelOrder(context.Background(), sold.Order.Id.Hex(), f.branch.Hex(), "manager", ""); !errors.As(err, &rejected) {
		t.Fatalf("a second cancel must be refused, got %v", err)
	}
	if err := f.ledger.CancelOrder(context.Background(), sold.Order.Id.Hex(), primitive.NewObjectID().Hex(), "manager", ""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another branch's Order: want ErrNotFound, got %v", err)
	}
}

func TestCancelLineSavedWithoutAStatusIsCancelled(t *testing.T) {
	f := newFixture(t)
	lot := f.stock(f.tab, 1, 0)
	order, line := primitive.NewObjectID(), primitive.NewObjectID()
	f.insert("orders", entities.Order{Id: order, BranchId: f.branch, Status: constant.CONFIRMED})
	if _, err := f.pos.Collection("order_items").InsertOne(context.Background(), bson.M{"_id": line, "orderId": order, "branchId": f.branch, "productId": f.product, "unitId": f.tab,
		"quantity": 2, "price": 20.0, "stocks": bson.A{bson.M{"stockid": lot.Hex(), "quantity": 2}}}); err != nil {
		t.Fatal(err)
	}
	if err := f.ledger.CancelLine(context.Background(), line.Hex(), f.branch.Hex(), "manager", ""); err != nil {
		t.Fatalf("an older Line with no status must cancel, got %v", err)
	}
	if f.line(line).Status != constant.CANCELLED || f.quantity(lot) != 2 {
		t.Fatal("Line not cancelled or Stock not put back")
	}
}

func TestCancelSkipsAStockDeletedSinceTheSale(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	kept, gone := f.stock(f.tab, 1, 1), f.stock(f.tab, 2, 1)
	sold, err := f.ledger.Sell(context.Background(), f.sale("s1", 50, request.SaleLine{Quantity: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.pos.Collection("product_stocks").DeleteOne(context.Background(), bson.M{"_id": gone}); err != nil {
		t.Fatal(err)
	}
	if err := f.ledger.CancelOrder(context.Background(), sold.Order.Id.Hex(), f.branch.Hex(), "manager", ""); err != nil {
		t.Fatalf("a Stock deleted by hand must not block the cancel, got %v", err)
	}
	if f.quantity(kept) != 1 {
		t.Fatal("the Stock that still exists did not get its quantity back")
	}
}

func TestSellChangedBuyerDetailsUnderTheSameIdIsADifferentSale(t *testing.T) {
	f := newFixture(t)
	f.priced(f.tab, 10)
	f.stock(f.tab, 1, 5)
	first := f.sale("s1", 50, request.SaleLine{Quantity: 1})
	first.BuyerIdCard = "1-1111-11111-11-1"
	if _, err := f.ledger.Sell(context.Background(), first); err != nil {
		t.Fatal(err)
	}
	corrected := f.sale("s1", 50, request.SaleLine{Quantity: 1})
	corrected.BuyerIdCard = "1-2222-22222-22-2"
	if _, err := f.ledger.Sell(context.Background(), corrected); !errors.Is(err, ErrSaleConflict) {
		t.Fatalf("a resend with different buyer details must not silently return the old record, got %v", err)
	}
}

// --- Transfer, manual Stock and set-quantity (PR-3) ---

func (f *fixture) transfer(to primitive.ObjectID, stock primitive.ObjectID, qty int) request.StockTransfer {
	return request.StockTransfer{FromBranchId: f.branch.Hex(), ToBranchId: to.Hex(), Code: "TR-1", CreatedBy: "admin",
		Items: []request.StockTransferItem{{ProductId: f.product.Hex(), StockId: stock.Hex(), Quantity: qty}}}
}

func TestRequestTransferReservesFromTheSourceWithHistory(t *testing.T) {
	f := newFixture(t)
	source := f.stock(f.tab, 1, 5)
	got, err := f.ledger.RequestTransfer(context.Background(), f.transfer(primitive.NewObjectID(), source, 3))
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "PENDING" || f.quantity(source) != 2 {
		t.Fatalf("status %s, source %d: want PENDING and 2 reserved out", got.Status, f.quantity(source))
	}
	if h := f.histories(); len(h) != 1 || h[0].Balance != 2 {
		t.Fatalf("want one history row for the reservation, got %+v", h)
	}
	var rejected *Rejected
	if _, err := f.ledger.RequestTransfer(context.Background(), f.transfer(primitive.NewObjectID(), source, 3)); !errors.As(err, &rejected) {
		t.Fatalf("reserving more than the source holds: want Rejected, got %v", err)
	}
}

func TestApproveTransferOpensStockAtTheDestinationAndSettlesItsLinesOnce(t *testing.T) {
	f := newFixture(t)
	dest := primitive.NewObjectID()
	source := f.stock(f.tab, 7, 5)
	f.insert("product_stocks", entities.ProductStock{Id: primitive.NewObjectID(), BranchId: dest, ProductId: f.product, UnitId: f.tab, Sequence: 2})
	waiting := f.owed(dest, f.tab, 1)
	sourceWaiting := f.owed(f.branch, f.tab, 1)
	tr, err := f.ledger.RequestTransfer(context.Background(), f.transfer(dest, source, 3))
	if err != nil {
		t.Fatal(err)
	}

	got, err := f.ledger.ApproveTransfer(context.Background(), tr.Id.Hex(), "manager")
	if err != nil {
		t.Fatal(err)
	}
	var opened entities.ProductStock
	if err := f.pos.Collection("product_stocks").FindOne(context.Background(), bson.M{"branchId": dest, "import": 3}).Decode(&opened); err != nil {
		t.Fatal(err)
	}
	if got.Status != "APPROVED" || opened.Quantity != 2 || opened.Sequence != 3 || f.line(waiting).OversoldQty != 0 {
		t.Fatalf("status %s, opened %+v, waiting owed %d", got.Status, opened, f.line(waiting).OversoldQty)
	}
	if f.line(sourceWaiting).OversoldQty != 1 {
		t.Fatal("the source branch's Line was settled by goods leaving it")
	}
	var rejected *Rejected
	if _, err := f.ledger.ApproveTransfer(context.Background(), tr.Id.Hex(), "manager"); !errors.As(err, &rejected) {
		t.Fatalf("a second approve must be refused, got %v", err)
	}
}

func TestRejectTransferPutsTheReservationBackAndSettlesTheSource(t *testing.T) {
	f := newFixture(t)
	source := f.stock(f.tab, 1, 3)
	tr, err := f.ledger.RequestTransfer(context.Background(), f.transfer(primitive.NewObjectID(), source, 3))
	if err != nil {
		t.Fatal(err)
	}
	waiting := f.owed(f.branch, f.tab, 1)
	got, err := f.ledger.RejectTransfer(context.Background(), tr.Id.Hex(), "manager")
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != "REJECTED" || f.quantity(source) != 2 || f.line(waiting).OversoldQty != 0 {
		t.Fatalf("status %s, source %d, owed %d: want REJECTED, 3 back less 1 to the waiting Line", got.Status, f.quantity(source), f.line(waiting).OversoldQty)
	}
}

func TestCreateStockSettlesWaitingLinesWithOneHistoryRow(t *testing.T) {
	f := newFixture(t)
	f.stock(f.box, 4, 0)
	waiting := f.owed(f.branch, f.box, 2)
	got, err := f.ledger.CreateStock(context.Background(), request.ProductStock{ProductId: f.product.Hex(), UnitId: f.box.Hex(), BranchId: f.branch.Hex(), Quantity: 5, LotNumber: "B1", UpdatedBy: "admin"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Quantity != 3 || got.Import != 5 || got.Sequence != 5 || f.line(waiting).OversoldQty != 0 {
		t.Fatalf("created %+v, owed %d", got, f.line(waiting).OversoldQty)
	}
	if h := f.histories(); len(h) != 1 || h[0].Import != 5 || h[0].Balance != 3 {
		t.Fatalf("want one history row, got %+v", h)
	}
	var rejected *Rejected
	if _, err := f.ledger.CreateStock(context.Background(), request.ProductStock{ProductId: f.product.Hex(), UnitId: primitive.NewObjectID().Hex(), BranchId: f.branch.Hex(), Quantity: 1}); !errors.As(err, &rejected) {
		t.Fatalf("a Unit of another Product: want Rejected, got %v", err)
	}
}

func TestSetQuantityIsAOneLineCount(t *testing.T) {
	f := newFixture(t)
	st := f.stock(f.tab, 1, 5)
	waiting := f.owed(f.branch, f.tab, 1)
	got, err := f.ledger.SetQuantity(context.Background(), st.Hex(), f.branch.Hex(), 8, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if got.Quantity != 7 || f.line(waiting).OversoldQty != 0 || f.count("stock_adjustments", bson.M{"stockId": st, "delta": 3}) != 1 {
		t.Fatalf("Stock %+v: want 8 set, 1 to the waiting Line, one Adjustment of +3", got)
	}
	var rejected *Rejected
	if _, err := f.ledger.SetQuantity(context.Background(), st.Hex(), f.branch.Hex(), -1, "admin"); !errors.As(err, &rejected) {
		t.Fatalf("a negative quantity: want Rejected, got %v", err)
	}
	if _, err := f.ledger.SetQuantity(context.Background(), st.Hex(), primitive.NewObjectID().Hex(), 1, "admin"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("another branch's Stock: want ErrNotFound, got %v", err)
	}
}

func TestDeleteStockOnlyWhenEmpty(t *testing.T) {
	f := newFixture(t)
	full, empty := f.stock(f.tab, 1, 2), f.stock(f.tab, 2, 0)
	var rejected *Rejected
	if _, err := f.ledger.DeleteStock(context.Background(), full.Hex(), f.branch.Hex(), "admin"); !errors.As(err, &rejected) {
		t.Fatalf("a Stock still holding goods: want Rejected, got %v", err)
	}
	if _, err := f.ledger.DeleteStock(context.Background(), empty.Hex(), f.branch.Hex(), "admin"); err != nil {
		t.Fatal(err)
	}
	if f.count("product_stocks", bson.M{"_id": empty}) != 0 || len(f.histories()) != 1 {
		t.Fatal("the empty Stock was not deleted with one history row")
	}
}

func TestDeleteStockRefusesAStockAPendingTransferReserved(t *testing.T) {
	f := newFixture(t)
	source := f.stock(f.tab, 1, 3)
	tr, err := f.ledger.RequestTransfer(context.Background(), f.transfer(primitive.NewObjectID(), source, 3))
	if err != nil {
		t.Fatal(err)
	}
	var rejected *Rejected
	if _, err := f.ledger.DeleteStock(context.Background(), source.Hex(), f.branch.Hex(), "admin"); !errors.As(err, &rejected) {
		t.Fatalf("a Stock emptied by a pending Transfer must not be deleted, got %v", err)
	}
	if _, err := f.ledger.RejectTransfer(context.Background(), tr.Id.Hex(), "manager"); err != nil || f.quantity(source) != 3 {
		t.Fatalf("the reservation must still come back, err=%v Stock=%d", err, f.quantity(source))
	}
}

func TestRequestTransferRefusesLinesWithoutAStockOrNamingOneTwice(t *testing.T) {
	f := newFixture(t)
	source := f.stock(f.tab, 1, 5)
	var rejected *Rejected
	twice := f.transfer(primitive.NewObjectID(), source, 1)
	twice.Items = append(twice.Items, request.StockTransferItem{ProductId: f.product.Hex(), StockId: source.Hex(), Quantity: 2})
	noStock := f.transfer(primitive.NewObjectID(), source, 1)
	noStock.Items[0].StockId = ""
	for name, form := range map[string]request.StockTransfer{"same Stock twice": twice, "no Stock": noStock} {
		if _, err := f.ledger.RequestTransfer(context.Background(), form); !errors.As(err, &rejected) {
			t.Fatalf("%s: want Rejected, got %v", name, err)
		}
	}
	if f.quantity(source) != 5 || f.count("stock_transfers", bson.M{}) != 0 {
		t.Fatal("a refused Transfer reserved Stock or was recorded")
	}
}

func TestApprovedTransferHistoryRecordsWhatCameIn(t *testing.T) {
	f := newFixture(t)
	dest := primitive.NewObjectID()
	source := f.stock(f.tab, 1, 4)
	tr, err := f.ledger.RequestTransfer(context.Background(), f.transfer(dest, source, 4))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.ledger.ApproveTransfer(context.Background(), tr.Id.Hex(), "manager"); err != nil {
		t.Fatal(err)
	}
	var in entities.ProductHistory
	if err := f.pos.Collection("product_histories").FindOne(context.Background(), bson.M{"branchId": dest}).Decode(&in); err != nil {
		t.Fatal(err)
	}
	if in.Import != 4 || in.Quantity != 4 {
		t.Fatalf("inbound Transfer history %+v, want Import and Quantity 4", in)
	}
}
