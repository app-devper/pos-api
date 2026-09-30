package repositories

import (
	"context"
	"testing"

	"pos/app/data/entities"
	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestRepairCancelledLineTotals(t *testing.T) {
	f := newSaleFixture(t)
	ctx := context.Background()
	order := func(status string, total, totalCost, discount float64, lines ...entities.OrderItem) primitive.ObjectID {
		id := primitive.NewObjectID()
		f.insert("orders", bson.M{"_id": id, "code": "O-" + id.Hex()[18:], "status": status, "total": total, "totalCost": totalCost, "discount": discount})
		for _, l := range lines {
			l.Id, l.OrderId = primitive.NewObjectID(), id
			f.insert("order_items", l)
		}
		return id
	}
	cancelled := entities.OrderItem{Status: constant.CANCELLED, Quantity: 2, Price: 20, CostPrice: 8, Discount: 1}
	kept := entities.OrderItem{Status: constant.CONFIRMED, Quantity: 3, Price: 24, CostPrice: 12, Discount: 0.5}

	// As the old cancel left it: price × quantity, cost × quantity, discount untouched.
	broken := order(constant.CONFIRMED, 72, 36, 3.5, cancelled, kept)
	right := order(constant.CONFIRMED, 22.5, 12, 1.5, cancelled, kept)
	voided := order(constant.CANCELLED, 72, 36, 3.5, cancelled, kept)
	untouched := order(constant.CONFIRMED, 999, 999, 0, kept)
	staleDiscount := order(constant.CONFIRMED, 22.5, 12, 3.5, cancelled, kept)

	stored := func(id primitive.ObjectID) entities.Order {
		var o entities.Order
		if err := f.pos.Collection("orders").FindOne(ctx, bson.M{"_id": id}).Decode(&o); err != nil {
			t.Fatal(err)
		}
		return o
	}

	report, err := RepairCancelledLineTotals(ctx, f.pos, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(report) != 2 || report[1].OrderId != staleDiscount || report[1].NewDiscount != 1.5 {
		t.Fatalf("report %+v", report)
	}
	if report[0].OrderId != broken || report[0].NewTotal != 22.5 || report[0].NewTotalCost != 12 || report[0].NewDiscount != 1.5 || report[0].Total != 72 {
		t.Fatalf("report %+v", report)
	}
	if stored(broken).Total != 72 {
		t.Fatal("a report must not write")
	}

	if _, err := RepairCancelledLineTotals(ctx, f.pos, true); err != nil {
		t.Fatal(err)
	}
	if o := stored(broken); o.Total != 22.5 || o.TotalCost != 12 || o.Discount != 1.5 {
		t.Fatalf("repaired %+v", o)
	}
	if stored(right).Total != 22.5 || stored(voided).Total != 72 || stored(untouched).Total != 999 {
		t.Fatal("repaired an Order it should have left alone")
	}
	if again, err := RepairCancelledLineTotals(ctx, f.pos, true); err != nil || len(again) != 0 {
		t.Fatalf("second run %+v %v", again, err)
	}
}
