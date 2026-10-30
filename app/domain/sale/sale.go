// Package sale prices a Sale's Lines and decides which Stock each draws from
// (CONTEXT: Sale, Line, Price type, Override, Stock, Sold first, Oversell).
// The till only previews; pos-api records the Order from what this package
// decides, using the Stock as it stands.
//
// Money on a recorded Line: price is the Line amount (unit price × quantity,
// before discount), costPrice the Line's cost, and discount is per unit. The
// customer pays price − discount × quantity.
package sale

import (
	"sort"

	"pos/app/data/entities"
)

// PriceTypeStock is the price type that charges the Stock's own price.
const PriceTypeStock = "Stock"

// Line is one product at one quantity as the till rang it up.
type Line struct {
	ProductId string
	UnitId    string
	Quantity  int
	// PriceType is the Customer type's price list, or PriceTypeStock.
	PriceType string
	// StockId is a batch the cashier chose to sell from first; empty for none.
	StockId       string
	Discount      float64 // per unit
	AllowOversell bool
}

// Catalog is what the shop holds for a Line's Unit in the selling branch.
type Catalog struct {
	UnitCost float64
	// Prices of the Unit, one per price list.
	Prices []entities.ProductPrice
	// Stocks of the Unit in the branch, in any order.
	Stocks []entities.ProductStock
}

// Part is a quantity drawn from one Stock; StockId "" is Sold first.
type Part struct {
	StockId  string
	Quantity int
}

// Rung is a Line as the shop records it.
type Rung struct {
	// PriceType actually charged: a missing price list falls back to the
	// Unit's first one, and no price list at all charges 0 with type "".
	PriceType string
	UnitPrice float64
	Amount    float64 // unit price × quantity, before discount
	Cost      float64 // the cost of what was drawn
	Discount  float64 // per unit, between 0 and the unit price
	Parts     []Part
	// Oversold is how much of the last Stock's Part no Stock held: that
	// Stock goes to zero and the Line owes the rest.
	Oversold int
}

// Paid is what the customer pays for the Line.
func (r Rung) Paid(quantity int) float64 { return Round(r.Amount - r.Discount*float64(quantity)) }

// Ring prices a Line and draws its quantity from the Catalog's Stock, as the
// till previews it:
//
//   - The Line sells from the chosen Stock, or else the first Stock in
//     selling order (by sequence) that holds any.
//   - A Stock price type charges that Stock's own price when it has one;
//     otherwise the Line's price list, the Unit's first price list when the
//     customer type has none, or 0.
//   - The quantity comes from that Stock, then the others holding any in
//     selling order. What no Stock covers is Sold first, unless the Line
//     allows Oversell and already drew from a Stock: that Stock owes it.
//   - Cost is each Stock's cost for what it gave (the Unit's cost when the
//     Stock has none, and for Sold first).
func Ring(l Line, c Catalog) Rung {
	stocks := append([]entities.ProductStock(nil), c.Stocks...)
	sort.SliceStable(stocks, func(i, j int) bool { return stocks[i].Sequence < stocks[j].Sequence })
	first := firstStock(l, stocks)

	r := Rung{}
	r.PriceType, r.UnitPrice = price(l, c.Prices, first)
	charge := ChargeFor(r.UnitPrice, l.Quantity, l.Discount)
	r.Amount, r.Discount = charge.Amount, charge.Discount
	r.Parts, r.Oversold = allocate(l, stocks, first)

	costs := map[string]float64{"": c.UnitCost}
	for _, s := range stocks {
		costs[s.Id.Hex()] = c.UnitCost
		if s.CostPrice > 0 {
			costs[s.Id.Hex()] = s.CostPrice
		}
	}
	for _, p := range r.Parts {
		r.Cost += costs[p.StockId] * float64(p.Quantity)
	}
	r.Cost = Round(r.Cost)
	return r
}

// firstStock is the chosen Stock, else the first in selling order holding any.
func firstStock(l Line, stocks []entities.ProductStock) *entities.ProductStock {
	if l.StockId != "" {
		for i := range stocks {
			if stocks[i].Id.Hex() == l.StockId {
				return &stocks[i]
			}
		}
	}
	for i := range stocks {
		if stocks[i].Quantity > 0 {
			return &stocks[i]
		}
	}
	return nil
}

func price(l Line, prices []entities.ProductPrice, stock *entities.ProductStock) (string, float64) {
	if l.PriceType == PriceTypeStock && stock != nil && stock.Price > 0 {
		return PriceTypeStock, stock.Price
	}
	if len(prices) == 0 {
		return "", 0
	}
	for _, p := range prices {
		if p.CustomerType == l.PriceType {
			return p.CustomerType, p.Price
		}
	}
	return prices[0].CustomerType, prices[0].Price
}

func allocate(l Line, stocks []entities.ProductStock, first *entities.ProductStock) ([]Part, int) {
	if l.Quantity <= 0 {
		return nil, 0
	}
	if first == nil {
		return []Part{{Quantity: l.Quantity}}, 0
	}
	var parts []Part
	remaining := l.Quantity
	take := func(s *entities.ProductStock) {
		q := min(max(s.Quantity, 0), remaining)
		parts = append(parts, Part{StockId: s.Id.Hex(), Quantity: q})
		remaining -= q
	}
	take(first)
	for i := range stocks {
		if remaining == 0 {
			break
		}
		if stocks[i].Id != first.Id {
			take(&stocks[i])
		}
	}
	oversold := 0
	if remaining > 0 {
		if l.AllowOversell {
			parts[len(parts)-1].Quantity += remaining
			oversold = remaining
		} else {
			parts = append(parts, Part{Quantity: remaining})
		}
	}
	// Stocks that hold none give nothing.
	kept := parts[:0]
	for _, p := range parts {
		if p.Quantity > 0 {
			kept = append(kept, p)
		}
	}
	return kept, oversold
}
