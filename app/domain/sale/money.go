package sale

import (
	"math"
	"pos/app/data/entities"
)

// Money on a Sale. pos-api owns it; the till previews with the same rules
// (devper-workspace: packages/applications/pos/lib/domain/model/sale/money.dart).
// testdata/money_cases.json holds the cases both sides must agree on, and a
// copy of it lives in the workspace's tests: change one, change both.

// Round rounds to the satang: half away from zero, on v×100 as a float64.
// The till rounds the same float the same way, which matters more than
// rounding exactly — 4.445 rounds to 4.44 on both.
func Round(v float64) float64 { return math.Round(v*100) / 100 }

// Charge is what one Line charges.
type Charge struct {
	Amount   float64 // unit price × quantity, rounded, before discount
	Discount float64 // per unit, between 0 and the unit price
	Paid     float64 // what the customer pays for the Line
}

// ChargeFor prices quantity units at unitPrice less discount per unit.
func ChargeFor(unitPrice float64, quantity int, discount float64) Charge {
	c := Charge{
		Amount:   Round(unitPrice * float64(quantity)),
		Discount: math.Min(math.Max(discount, 0), unitPrice),
	}
	c.Paid = Round(c.Amount - c.Discount*float64(quantity))
	return c
}

// Totals sums an Order's Lines: each Line's paid amount is already rounded,
// and the sums are rounded once.
type Totals struct{ total, cost, discount float64 }

// Add counts one Line: what it was paid, what it cost, and its whole discount.
func (t *Totals) Add(paid, cost, discount float64) {
	t.total += paid
	t.cost += cost
	t.discount += discount
}

// Rounded is the Order's total, total cost and discount.
func (t Totals) Rounded() (total, cost, discount float64) {
	return Round(t.total), Round(t.cost), Round(t.discount)
}

// Tender reports whether the money offered settles total, and the change.
// Change comes from the tender rounded to the satang: 8.895 offered against
// 8.90 settles it, and must not record −0.01 change.
func Tender(tendered, total float64) (change float64, covers bool) {
	offered := Round(tendered)
	if offered < total {
		return 0, false
	}
	return Round(offered - total), true
}

// OrderMoney is an Order's total, total cost and discount from its standing
// Lines, the way a Sale summed them: a Line's price is its amount before
// discount and its discount is per unit, so each pays Round(price −
// discount × quantity). A cancel and the repair command recompute an Order
// with it.
func OrderMoney(lines []entities.OrderItem) (total, cost, discount float64) {
	var totals Totals
	for _, line := range lines {
		qty := float64(line.Quantity)
		totals.Add(Round(line.Price-line.Discount*qty), line.CostPrice, line.Discount*qty)
	}
	return totals.Rounded()
}
