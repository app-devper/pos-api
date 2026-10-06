package repositories

import (
	"context"
	"sync"
	"testing"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func stockCommands(f *saleFixture) (IProductReturn, IStockAdjustment, IStockCount) {
	resource := &db.Resource{Client: f.pos.Client(), PosDb: f.pos}
	NewProductStockEntity(resource)
	NewSequenceEntity(resource)
	return NewProductReturnEntity(resource), NewStockAdjustmentEntity(resource), NewStockCountEntity(resource)
}

func returnRequest(f *saleFixture, order primitive.ObjectID, item primitive.ObjectID, qty int, refund float64) request.ProductReturn {
	return request.ProductReturn{OrderId: order.Hex(), BranchId: f.branchId.Hex(), CreatedBy: "admin", Items: []request.ProductReturnItem{{OrderItemId: item.Hex(), Quantity: qty, Refund: refund}}}
}

func rejectInserts(t *testing.T, f *saleFixture, collection string) {
	t.Helper()
	// Fails at the final persistence step, after the command has changed Stock.
	err := f.pos.RunCommand(context.Background(), bson.D{{Key: "collMod", Value: collection}, {Key: "validator", Value: bson.M{"_id": bson.M{"$exists": false}}}, {Key: "validationLevel", Value: "strict"}}).Err()
	if err != nil {
		t.Fatal(err)
	}
}

func returnedQuantity(f *saleFixture, item primitive.ObjectID) int {
	f.t.Helper()
	var got entities.OrderItem
	if err := f.pos.Collection("order_items").FindOne(context.Background(), bson.M{"_id": item}).Decode(&got); err != nil {
		f.t.Fatal(err)
	}
	return got.ReturnedQty
}

func TestReturnRecordingRestoresOriginalLotsAndCapsRefund(t *testing.T) {
	f := newSaleFixture(t)
	returns, _, _ := stockCommands(f)
	first := f.stock(1, 3, 3)
	second := f.stock(2, 3, 3)
	order, err := f.orders.RecordSale(f.sale("r1", 100, request.SaleLine{Quantity: 5, PriceType: "General", Discount: 1}))
	if err != nil {
		t.Fatal(err)
	}
	item := f.items(order.Order.Id)[0]
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 2, 18)); err != nil {
		t.Fatal(err)
	}
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 2, 18)); err != nil {
		t.Fatal(err)
	}
	if f.quantity(first) != 3 || f.quantity(second) != 2 || returnedQuantity(f, item.Id) != 4 {
		t.Fatal("partial Return did not resume at the original Lot")
	}
	beforeHistory := f.count("product_histories")
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 1, 9.02)); err == nil {
		t.Fatal("refund above paid share accepted")
	}
	if f.quantity(second) != 2 || returnedQuantity(f, item.Id) != 4 || f.count("product_histories") != beforeHistory {
		t.Fatal("rejected refund changed Stock or history")
	}
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 2, 0)); err == nil {
		t.Fatal("returned more than sold")
	}
}

func TestReturnRecordingRollsBackWhenDocumentFails(t *testing.T) {
	f := newSaleFixture(t)
	returns, _, _ := stockCommands(f)
	lot := f.stock(1, 5, 3)
	order, err := f.orders.RecordSale(f.sale("r2", 100, request.SaleLine{Quantity: 3, PriceType: "General"}))
	if err != nil {
		t.Fatal(err)
	}
	item := f.items(order.Order.Id)[0]
	history := f.count("product_histories")
	rejectInserts(t, f, "product_returns")
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 2, 20)); err == nil {
		t.Fatal("expected persistence failure")
	}
	if f.quantity(lot) != 2 || returnedQuantity(f, item.Id) != 0 || f.count("product_returns") != 0 || f.count("product_histories") != history {
		t.Fatal("Return left partial effects")
	}
}

func TestReturnRecordingRejectsDuplicateAndCancelledLines(t *testing.T) {
	f := newSaleFixture(t)
	returns, _, _ := stockCommands(f)
	lot := f.stock(1, 5, 3)
	order, err := f.orders.RecordSale(f.sale("r3", 100, request.SaleLine{Quantity: 3, PriceType: "General"}))
	if err != nil {
		t.Fatal(err)
	}
	item := f.items(order.Order.Id)[0]
	req := returnRequest(f, order.Order.Id, item.Id, 2, 0)
	req.Items = append(req.Items, req.Items[0])
	if _, err := returns.RecordProductReturn(req); err == nil {
		t.Fatal("duplicate Line accepted")
	}
	if f.quantity(lot) != 2 || returnedQuantity(f, item.Id) != 0 {
		t.Fatal("duplicate Line changed Stock")
	}
	if _, err := f.orders.CancelOrderItemById(item.Id.Hex(), "admin", f.branchId.Hex(), "cancel"); err != nil {
		t.Fatal(err)
	}
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 1, 0)); err == nil {
		t.Fatal("cancelled Line returned twice")
	}
	if f.quantity(lot) != 5 {
		t.Fatal("cancelled Line restored Stock twice")
	}
}

func TestConcurrentReturnsCannotExceedRealLotQuantity(t *testing.T) {
	f := newSaleFixture(t)
	returns, _, _ := stockCommands(f)
	lot := f.stock(1, 5, 3)
	order, err := f.orders.RecordSale(f.sale("r4", 100, request.SaleLine{Quantity: 5, PriceType: "General"}))
	if err != nil {
		t.Fatal(err)
	}
	item := f.items(order.Order.Id)[0]
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 4, 0))
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		}
	}
	if success != 1 || f.quantity(lot) != 4 || returnedQuantity(f, item.Id) != 4 || f.count("product_returns") != 1 {
		t.Fatalf("concurrent Return duplicated Stock: successes=%d", success)
	}
}

func adjustmentRequest(f *saleFixture, stock primitive.ObjectID, delta int) request.StockAdjustment {
	return request.StockAdjustment{BranchId: f.branchId.Hex(), ProductId: f.product.Hex(), StockId: stock.Hex(), Reason: constant.AdjustmentReasonCount, Delta: delta, CreatedBy: "admin"}
}

func TestAdjustmentRollsBackStockHistoryAndReconciliation(t *testing.T) {
	f := newSaleFixture(t)
	_, adjustments, _ := stockCommands(f)
	stock := f.stock(1, 0, 3)
	item := primitive.NewObjectID()
	f.insert("order_items", entities.OrderItem{Id: item, ProductId: f.product, BranchId: f.branchId, Status: constant.CONFIRMED, OversoldQty: 2})
	rejectInserts(t, f, "stock_adjustments")
	if _, err := adjustments.ApplyStockAdjustment(adjustmentRequest(f, stock, 3)); err == nil {
		t.Fatal("expected failed Adjustment document")
	}
	if f.quantity(stock) != 0 || f.count("product_histories") != 0 || f.count("stock_adjustments") != 0 {
		t.Fatal("Adjustment left partial effects")
	}
	var got entities.OrderItem
	if err := f.pos.Collection("order_items").FindOne(context.Background(), bson.M{"_id": item}).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.OversoldQty != 2 {
		t.Fatal("failed Adjustment changed reconciliation")
	}
}

func TestAdjustmentSettlesFIFOFromTheStockAndPreservesValidation(t *testing.T) {
	f := newSaleFixture(t)
	_, adjustments, _ := stockCommands(f)
	stock := f.stock(1, 2, 3)
	first, second := primitive.NewObjectID(), primitive.NewObjectID()
	for _, id := range []primitive.ObjectID{first, second} {
		f.insert("order_items", entities.OrderItem{Id: id, ProductId: f.product, UnitId: f.unit, BranchId: f.branchId, Status: constant.CONFIRMED, OversoldQty: 2})
	}
	got, err := adjustments.ApplyStockAdjustment(adjustmentRequest(f, stock, 3))
	if err != nil {
		t.Fatal(err)
	}
	// ADR-0002: the 3 that came in go to the waiting Lines, oldest first.
	if got.Before != 2 || got.After != 5 || f.quantity(stock) != 2 {
		t.Fatalf("Adjustment %d → %d, Stock now %d; want 2 → 5 and Stock 2", got.Before, got.After, f.quantity(stock))
	}
	for i, want := range []struct {
		id    primitive.ObjectID
		owed  int
		drawn int
	}{{first, 0, 2}, {second, 1, 1}} {
		var item entities.OrderItem
		if err := f.pos.Collection("order_items").FindOne(context.Background(), bson.M{"_id": want.id}).Decode(&item); err != nil {
			t.Fatal(err)
		}
		if item.OversoldQty != want.owed || len(item.Stocks) != 1 || item.Stocks[0] != (entities.OrderItemStock{StockId: stock.Hex(), Quantity: want.drawn}) {
			t.Fatalf("Line %d not settled FIFO from the Stock: %+v", i, item)
		}
	}
	for _, delta := range []int{0, -6} {
		if _, err := adjustments.ApplyStockAdjustment(adjustmentRequest(f, stock, delta)); err == nil {
			t.Fatal("invalid delta accepted")
		}
	}
	req := adjustmentRequest(f, stock, 1)
	req.BranchId = primitive.NewObjectID().Hex()
	if _, err := adjustments.ApplyStockAdjustment(req); err == nil {
		t.Fatal("another branch's Stock adjusted")
	}
	req = adjustmentRequest(f, stock, 1)
	req.ProductId = primitive.NewObjectID().Hex()
	if _, err := adjustments.ApplyStockAdjustment(req); err == nil {
		t.Fatal("Stock attributed to another Product")
	}
	if f.quantity(stock) != 2 || f.count("stock_adjustments") != 1 {
		t.Fatal("rejected Adjustment changed Stock")
	}
}

func TestCountRecordsAllLinesAndOnlyChangedAdjustments(t *testing.T) {
	f := newSaleFixture(t)
	_, _, counts := stockCommands(f)
	first, second := f.stock(1, 5, 3), f.stock(2, 5, 3)
	result, err := counts.RecordStockCount(request.StockCount{BranchId: f.branchId.Hex(), CreatedBy: "admin", Items: []request.StockCountItem{{ProductId: f.product.Hex(), StockId: first.Hex(), Counted: 5}, {ProductId: f.product.Hex(), StockId: second.Hex(), Counted: 8}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 2 || result.Items[0].Delta != 0 || result.Items[1].SystemQuantity != 5 || result.Items[1].Delta != 3 || f.count("stock_adjustments") != 1 || f.quantity(second) != 8 {
		t.Fatalf("Count outcome incorrect: %+v", result)
	}
}

func TestCountRollsBackEarlierLinesAndFinalDocumentFailure(t *testing.T) {
	for _, failure := range []string{"later Line", "final document"} {
		t.Run(failure, func(t *testing.T) {
			f := newSaleFixture(t)
			_, _, counts := stockCommands(f)
			first := f.stock(1, 5, 3)
			req := request.StockCount{BranchId: f.branchId.Hex(), Items: []request.StockCountItem{{ProductId: f.product.Hex(), StockId: first.Hex(), Counted: 8}}}
			if failure == "later Line" {
				req.Items = append(req.Items, request.StockCountItem{ProductId: f.product.Hex(), StockId: primitive.NewObjectID().Hex(), Counted: 4})
			} else {
				rejectInserts(t, f, "stock_counts")
			}
			if _, err := counts.RecordStockCount(req); err == nil {
				t.Fatal("expected Count failure")
			}
			if f.quantity(first) != 5 || f.count("stock_adjustments") != 0 || f.count("product_histories") != 0 || f.count("stock_counts") != 0 || f.count("sequences") != 0 {
				t.Fatal("failed Count left partial effects")
			}
		})
	}
}

func TestCountValidatesUnchangedLinesAndDuplicateStock(t *testing.T) {
	f := newSaleFixture(t)
	_, _, counts := stockCommands(f)
	stock := f.stock(1, 5, 3)
	line := request.StockCountItem{ProductId: f.product.Hex(), StockId: stock.Hex(), Counted: 5}
	for _, test := range []string{"branch", "Product", "duplicate", "negative"} {
		req := request.StockCount{BranchId: f.branchId.Hex(), Items: []request.StockCountItem{line}}
		switch test {
		case "branch":
			req.BranchId = primitive.NewObjectID().Hex()
		case "Product":
			req.Items[0].ProductId = primitive.NewObjectID().Hex()
		case "duplicate":
			req.Items = append(req.Items, line)
		case "negative":
			req.Items[0].Counted = -1
		}
		if _, err := counts.RecordStockCount(req); err == nil {
			t.Fatalf("accepted invalid Count: %s", test)
		}
	}
	if f.quantity(stock) != 5 || f.count("stock_counts") != 0 {
		t.Fatal("rejected unchanged Count changed Stock or recorded document")
	}
}

func TestConcurrentAdjustmentsKeepBeforeAfterConsistent(t *testing.T) {
	f := newSaleFixture(t)
	_, adjustments, _ := stockCommands(f)
	stock := f.stock(1, 5, 3)
	results := make(chan *entities.StockAdjustment, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := adjustments.ApplyStockAdjustment(adjustmentRequest(f, stock, 1))
			results <- result
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	before := map[int]bool{}
	for result := range results {
		before[result.Before] = true
		if result.After != result.Before+1 {
			t.Fatal("inconsistent Adjustment delta")
		}
	}
	if !before[5] || !before[6] || f.quantity(stock) != 7 {
		t.Fatal("concurrent Adjustment lost Stock or used stale before quantity")
	}
}

func TestCancelAfterReturnRestoresOnlyRemainingQuantity(t *testing.T) {
	f := newSaleFixture(t)
	returns, _, _ := stockCommands(f)
	first, second := f.stock(1, 3, 3), f.stock(2, 3, 3)
	order, err := f.orders.RecordSale(f.sale("return-then-cancel", 100, request.SaleLine{Quantity: 5, PriceType: "General"}))
	if err != nil {
		t.Fatal(err)
	}
	item := f.items(order.Order.Id)[0]
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 4, 0)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.orders.CancelOrderById(order.Order.Id.Hex(), "admin", f.branchId.Hex(), "cancel remainder"); err != nil {
		t.Fatal(err)
	}
	if f.quantity(first) != 3 || f.quantity(second) != 3 {
		t.Fatal("cancellation restored the returned quantity twice")
	}
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 1, 0)); err == nil {
		t.Fatal("cancelled Order accepted a Return")
	}
}

func TestReturnHistoryFailureRollsBackStockAndReturnedQuantity(t *testing.T) {
	f := newSaleFixture(t)
	returns, _, _ := stockCommands(f)
	stock := f.stock(1, 5, 3)
	order, err := f.orders.RecordSale(f.sale("history-fails", 100, request.SaleLine{Quantity: 3, PriceType: "General"}))
	if err != nil {
		t.Fatal(err)
	}
	item := f.items(order.Order.Id)[0]
	history := f.count("product_histories")
	rejectInserts(t, f, "product_histories")
	if _, err := returns.RecordProductReturn(returnRequest(f, order.Order.Id, item.Id, 1, 0)); err == nil {
		t.Fatal("ignored Return history failure")
	}
	if f.quantity(stock) != 2 || returnedQuantity(f, item.Id) != 0 || f.count("product_returns") != 0 || f.count("product_histories") != history {
		t.Fatal("history failure left Return effects")
	}
}

func TestReturnCannotRestoreSoldFirstOrSyntheticAllocations(t *testing.T) {
	for _, ref := range []string{"", "ADJUST:นับสต็อก"} {
		t.Run(ref, func(t *testing.T) {
			f := newSaleFixture(t)
			returns, _, _ := stockCommands(f)
			order, item := primitive.NewObjectID(), primitive.NewObjectID()
			f.insert("orders", entities.Order{Id: order, BranchId: f.branchId, Status: constant.CONFIRMED})
			f.insert("order_items", entities.OrderItem{Id: item, OrderId: order, BranchId: f.branchId, ProductId: f.product, UnitId: f.unit, Quantity: 3, Price: 30, Stocks: []entities.OrderItemStock{{StockId: ref, Quantity: 3}}})
			if _, err := returns.RecordProductReturn(returnRequest(f, order, item, 1, 0)); err == nil {
				t.Fatal("returned quantity without a real Lot")
			}
			if returnedQuantity(f, item) != 0 || f.count("product_returns") != 0 || f.count("product_histories") != 0 {
				t.Fatal("rejected Return left effects")
			}
		})
	}
}

func TestCountAndAdjustmentSerializeTheirStockSnapshot(t *testing.T) {
	f := newSaleFixture(t)
	_, adjustments, counts := stockCommands(f)
	stock := f.stock(1, 5, 3)
	start := make(chan struct{})
	done := make(chan error, 2)
	var count *entities.StockCount
	var adjustment *entities.StockAdjustment
	go func() {
		<-start
		var err error
		count, err = counts.RecordStockCount(request.StockCount{BranchId: f.branchId.Hex(), Items: []request.StockCountItem{{ProductId: f.product.Hex(), StockId: stock.Hex(), Counted: 10}}})
		done <- err
	}()
	go func() {
		<-start
		var err error
		adjustment, err = adjustments.ApplyStockAdjustment(adjustmentRequest(f, stock, 1))
		done <- err
	}()
	close(start)
	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	system := count.Items[0].SystemQuantity
	switch system {
	case 5:
		if adjustment.Before != 10 || f.quantity(stock) != 11 {
			t.Fatal("Adjustment did not observe committed Count")
		}
	case 6:
		if adjustment.Before != 5 || f.quantity(stock) != 10 {
			t.Fatal("Count used delta from an old snapshot")
		}
	default:
		t.Fatalf("inconsistent Count snapshot: %d", system)
	}
	if count.Items[0].Delta != 10-system || f.count("stock_adjustments") != 2 {
		t.Fatal("Count and Adjustment audit diverged")
	}
}
