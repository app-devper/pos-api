// Package stock holds the Stock rules that need no database: which waiting
// Lines incoming Stock serves, and which Stocks a Return goes back into
// (CONTEXT: Stock, Oversell, Settle, Return; ADR-0002).
package stock

import (
	"strings"

	"pos/app/data/entities"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SyntheticPrefix marks an allocation written by the old Adjustment
// settlement, which named no Stock. Nothing writes it any more.
const SyntheticPrefix = "ADJUST:"

// Debt is what one oversold Line is still owed.
type Debt struct {
	Line primitive.ObjectID
	Owed int
}

// Settlement is the part of incoming Stock that goes to one waiting Line.
type Settlement struct {
	Line primitive.ObjectID
	Qty  int
}

// Settle serves debts in the order given (oldest Line first) out of the
// incoming quantity, until either runs out.
func Settle(incoming int, debts []Debt) []Settlement {
	var settled []Settlement
	for _, d := range debts {
		if incoming <= 0 {
			break
		}
		if d.Owed <= 0 {
			continue
		}
		qty := min(incoming, d.Owed)
		settled = append(settled, Settlement{Line: d.Line, Qty: qty})
		incoming -= qty
	}
	return settled
}

// IsRealStock reports whether an allocation names a Stock, rather than Sold
// first ("") or an old synthetic Adjustment marker.
func IsRealStock(stockID string) bool {
	return stockID != "" && !strings.HasPrefix(stockID, SyntheticPrefix)
}

// RealQuantity is how much of a Line came out of named Stocks.
func RealQuantity(parts []entities.OrderItemStock) int {
	total := 0
	for _, p := range parts {
		if IsRealStock(p.StockId) {
			total += p.Quantity
		}
	}
	return total
}

// AllocateReturn picks the Stocks a Return of qty goes back into: the Line's
// named Stocks in the order it drew them, skipping what earlier Returns
// already put back.
func AllocateReturn(parts []entities.OrderItemStock, alreadyReturned, qty int) []entities.OrderItemStock {
	allocations := make([]entities.OrderItemStock, 0, len(parts))
	skip := alreadyReturned
	for _, p := range parts {
		if qty <= 0 {
			break
		}
		if !IsRealStock(p.StockId) {
			continue
		}
		available := p.Quantity
		if skip > 0 {
			used := min(skip, available)
			skip -= used
			available -= used
		}
		take := min(available, qty)
		if take <= 0 {
			continue
		}
		allocations = append(allocations, entities.OrderItemStock{StockId: p.StockId, Quantity: take})
		qty -= take
	}
	return allocations
}
