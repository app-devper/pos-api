package usecase

import (
	"testing"

	"pos/app/data/entities"
)

func TestRealLotQuantityExcludesSyntheticStockRefs(t *testing.T) {
	stocks := []entities.OrderItemStock{
		{StockId: "lot-1", Quantity: 3},
		{StockId: "ADJUST:นับสต็อก", Quantity: 2},
		{StockId: "lot-2", Quantity: 4},
	}
	if got := realLotQuantity(stocks); got != 7 {
		t.Fatalf("expected real lot quantity 7, got %d", got)
	}
}

func TestAllocateReturnAcrossRealLotsSkipsAlreadyReturnedAndSynthetic(t *testing.T) {
	stocks := []entities.OrderItemStock{
		{StockId: "lot-1", Quantity: 3},
		{StockId: "ADJUST:สูญหาย", Quantity: 5},
		{StockId: "lot-2", Quantity: 4},
	}

	allocations := allocateReturnAcrossRealLots(stocks, 2, 4)

	if len(allocations) != 2 {
		t.Fatalf("expected 2 allocations, got %+v", allocations)
	}
	if allocations[0].StockId != "lot-1" || allocations[0].Quantity != 1 {
		t.Fatalf("expected 1 unit remaining from lot-1 after skipping 2 already returned, got %+v", allocations[0])
	}
	if allocations[1].StockId != "lot-2" || allocations[1].Quantity != 3 {
		t.Fatalf("expected 3 units from lot-2 to complete the return of 4, got %+v", allocations[1])
	}
}

func TestAllocateReturnAcrossRealLotsCapsAtAvailable(t *testing.T) {
	stocks := []entities.OrderItemStock{
		{StockId: "lot-1", Quantity: 2},
	}

	allocations := allocateReturnAcrossRealLots(stocks, 0, 5)

	total := 0
	for _, a := range allocations {
		total += a.Quantity
	}
	if total != 2 {
		t.Fatalf("expected allocations capped at available quantity 2, got total %d (%+v)", total, allocations)
	}
}
