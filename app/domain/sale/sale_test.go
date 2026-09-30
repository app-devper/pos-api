package sale

import (
	"reflect"
	"testing"

	"pos/app/data/entities"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func stock(seq, qty int, cost, price float64) entities.ProductStock {
	return entities.ProductStock{Id: primitive.NewObjectID(), Sequence: seq, Quantity: qty, CostPrice: cost, Price: price}
}

func prices(pairs ...any) []entities.ProductPrice {
	var out []entities.ProductPrice
	for i := 0; i < len(pairs); i += 2 {
		out = append(out, entities.ProductPrice{CustomerType: pairs[i].(string), Price: pairs[i+1].(float64)})
	}
	return out
}

func TestRingDrawsInSellingOrderThenSoldFirst(t *testing.T) {
	a, b := stock(2, 3, 12, 0), stock(1, 2, 10, 0)
	r := Ring(Line{Quantity: 7, PriceType: "General"}, Catalog{UnitCost: 9, Prices: prices("General", 20.0), Stocks: []entities.ProductStock{a, b}})

	want := []Part{{b.Id.Hex(), 2}, {a.Id.Hex(), 3}, {"", 2}}
	if !reflect.DeepEqual(r.Parts, want) || r.Oversold != 0 {
		t.Fatalf("parts %+v oversold %d", r.Parts, r.Oversold)
	}
	if r.Amount != 140 || r.UnitPrice != 20 {
		t.Fatalf("amount %v unit price %v", r.Amount, r.UnitPrice)
	}
	if r.Cost != 2*10+3*12+2*9 {
		t.Fatalf("cost %v", r.Cost)
	}
}

func TestRingDrawsFromChosenStockFirst(t *testing.T) {
	a, b := stock(1, 5, 10, 0), stock(2, 5, 11, 0)
	r := Ring(Line{Quantity: 6, StockId: b.Id.Hex()}, Catalog{Stocks: []entities.ProductStock{a, b}})
	want := []Part{{b.Id.Hex(), 5}, {a.Id.Hex(), 1}}
	if !reflect.DeepEqual(r.Parts, want) {
		t.Fatalf("parts %+v", r.Parts)
	}
}

func TestRingIgnoresChosenStockOfAnotherUnit(t *testing.T) {
	a := stock(1, 5, 10, 0)
	r := Ring(Line{Quantity: 1, StockId: primitive.NewObjectID().Hex()}, Catalog{Stocks: []entities.ProductStock{a}})
	if !reflect.DeepEqual(r.Parts, []Part{{a.Id.Hex(), 1}}) {
		t.Fatalf("parts %+v", r.Parts)
	}
}

func TestRingOversellFoldsShortfallOntoLastStock(t *testing.T) {
	a, b := stock(1, 2, 10, 0), stock(2, 1, 11, 0)
	r := Ring(Line{Quantity: 5, AllowOversell: true}, Catalog{Stocks: []entities.ProductStock{a, b}})
	want := []Part{{a.Id.Hex(), 2}, {b.Id.Hex(), 3}}
	if !reflect.DeepEqual(r.Parts, want) || r.Oversold != 2 {
		t.Fatalf("parts %+v oversold %d", r.Parts, r.Oversold)
	}
	if r.Cost != 2*10+3*11 {
		t.Fatalf("cost %v", r.Cost)
	}
}

func TestRingOversellOntoChosenEmptyStock(t *testing.T) {
	a := stock(1, 0, 10, 0)
	r := Ring(Line{Quantity: 3, StockId: a.Id.Hex(), AllowOversell: true}, Catalog{Stocks: []entities.ProductStock{a}})
	if !reflect.DeepEqual(r.Parts, []Part{{a.Id.Hex(), 3}}) || r.Oversold != 3 {
		t.Fatalf("parts %+v oversold %d", r.Parts, r.Oversold)
	}
}

func TestRingWithoutStockIsSoldFirstEvenWithOversell(t *testing.T) {
	r := Ring(Line{Quantity: 4, AllowOversell: true}, Catalog{UnitCost: 7, Stocks: []entities.ProductStock{stock(1, 0, 10, 0)}})
	if !reflect.DeepEqual(r.Parts, []Part{{"", 4}}) || r.Oversold != 0 || r.Cost != 28 {
		t.Fatalf("parts %+v oversold %d cost %v", r.Parts, r.Oversold, r.Cost)
	}
}

func TestRingPrices(t *testing.T) {
	withPrice := stock(1, 5, 10, 33)
	noPrice := stock(1, 5, 10, 0)
	list := prices("General", 20.0, "Wholesaler", 15.0)
	for name, tc := range map[string]struct {
		line      Line
		catalog   Catalog
		priceType string
		price     float64
	}{
		"price list":             {Line{Quantity: 1, PriceType: "Wholesaler"}, Catalog{Prices: list}, "Wholesaler", 15},
		"missing list, first":    {Line{Quantity: 1, PriceType: "Regular"}, Catalog{Prices: list}, "General", 20},
		"no price list":          {Line{Quantity: 1, PriceType: "General"}, Catalog{}, "", 0},
		"stock price":            {Line{Quantity: 1, PriceType: PriceTypeStock}, Catalog{Prices: list, Stocks: []entities.ProductStock{withPrice}}, PriceTypeStock, 33},
		"stock without price":    {Line{Quantity: 1, PriceType: PriceTypeStock}, Catalog{Prices: list, Stocks: []entities.ProductStock{noPrice}}, "General", 20},
		"stock price, no stocks": {Line{Quantity: 1, PriceType: PriceTypeStock}, Catalog{Prices: list}, "General", 20},
	} {
		r := Ring(tc.line, tc.catalog)
		if r.PriceType != tc.priceType || r.UnitPrice != tc.price {
			t.Errorf("%s: got %q %v", name, r.PriceType, r.UnitPrice)
		}
	}
}

func TestRingStockPriceUsesChosenStock(t *testing.T) {
	a, b := stock(1, 5, 10, 30), stock(2, 5, 10, 25)
	r := Ring(Line{Quantity: 1, PriceType: PriceTypeStock, StockId: b.Id.Hex()}, Catalog{Stocks: []entities.ProductStock{a, b}})
	if r.UnitPrice != 25 {
		t.Fatalf("unit price %v", r.UnitPrice)
	}
}

func TestRingClampsDiscount(t *testing.T) {
	c := Catalog{Prices: prices("General", 20.0)}
	for discount, want := range map[float64]float64{-5: 0, 3: 3, 25: 20} {
		r := Ring(Line{Quantity: 2, PriceType: "General", Discount: discount}, c)
		if r.Discount != want {
			t.Errorf("discount %v: got %v", discount, r.Discount)
		}
	}
	r := Ring(Line{Quantity: 3, PriceType: "General", Discount: 2.5}, c)
	if r.Paid(3) != 52.5 {
		t.Fatalf("paid %v", r.Paid(3))
	}
}

func TestRingSkipsEmptyStocksAndUsesUnitCostWhenStockHasNone(t *testing.T) {
	empty, a := stock(1, 0, 10, 0), stock(2, 4, 0, 0)
	r := Ring(Line{Quantity: 2}, Catalog{UnitCost: 6, Stocks: []entities.ProductStock{empty, a}})
	if !reflect.DeepEqual(r.Parts, []Part{{a.Id.Hex(), 2}}) || r.Cost != 12 {
		t.Fatalf("parts %+v cost %v", r.Parts, r.Cost)
	}
}

func TestRingChosenEmptyStockGivesNothing(t *testing.T) {
	empty, a := stock(1, 0, 10, 0), stock(2, 5, 11, 0)
	r := Ring(Line{Quantity: 2, StockId: empty.Id.Hex()}, Catalog{Stocks: []entities.ProductStock{empty, a}})
	if !reflect.DeepEqual(r.Parts, []Part{{a.Id.Hex(), 2}}) {
		t.Fatalf("parts %+v", r.Parts)
	}
}
