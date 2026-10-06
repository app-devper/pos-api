package stock

import (
	"reflect"
	"testing"

	"pos/app/data/entities"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSettleServesTheOldestLinesFirst(t *testing.T) {
	a, b, c := primitive.NewObjectID(), primitive.NewObjectID(), primitive.NewObjectID()
	debts := []Debt{{Line: a, Owed: 2}, {Line: b, Owed: 3}, {Line: c, Owed: 1}}

	cases := []struct {
		name     string
		incoming int
		want     []Settlement
	}{
		{"nothing arrives", 0, nil},
		{"less than the first debt", 1, []Settlement{{Line: a, Qty: 1}}},
		{"spans two debts", 4, []Settlement{{Line: a, Qty: 2}, {Line: b, Qty: 2}}},
		{"more than all debts", 10, []Settlement{{Line: a, Qty: 2}, {Line: b, Qty: 3}, {Line: c, Qty: 1}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Settle(tc.incoming, debts); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Settle(%d) = %+v, want %+v", tc.incoming, got, tc.want)
			}
		})
	}
}

func TestSettleSkipsLinesThatOweNothing(t *testing.T) {
	a, b := primitive.NewObjectID(), primitive.NewObjectID()
	got := Settle(3, []Debt{{Line: a, Owed: 0}, {Line: b, Owed: 2}})
	if want := []Settlement{{Line: b, Qty: 2}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestRealQuantityCountsOnlyAllocationsThatNameAStock(t *testing.T) {
	parts := []entities.OrderItemStock{
		{StockId: primitive.NewObjectID().Hex(), Quantity: 3},
		{StockId: "", Quantity: 2},
		{StockId: "ADJUST:นับสต็อก", Quantity: 1},
	}
	if got := RealQuantity(parts); got != 3 {
		t.Fatalf("RealQuantity = %d, want 3", got)
	}
}

func TestAllocateReturnResumesAfterWhatWasAlreadyReturned(t *testing.T) {
	first, second := primitive.NewObjectID().Hex(), primitive.NewObjectID().Hex()
	parts := []entities.OrderItemStock{
		{StockId: first, Quantity: 3},
		{StockId: "ADJUST:อื่นๆ", Quantity: 4},
		{StockId: second, Quantity: 3},
	}
	got := AllocateReturn(parts, 2, 3)
	want := []entities.OrderItemStock{{StockId: first, Quantity: 1}, {StockId: second, Quantity: 2}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestAllocateReturnNeverExceedsWhatTheLineDrew(t *testing.T) {
	only := primitive.NewObjectID().Hex()
	got := AllocateReturn([]entities.OrderItemStock{{StockId: only, Quantity: 2}}, 0, 5)
	if want := []entities.OrderItemStock{{StockId: only, Quantity: 2}}; !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}
