// Package promotion decides what a Promotion is allowed to be and what it
// takes off a Sale. The handlers and the repository only carry it.
package promotion

import (
	"errors"
	"strings"

	"pos/app/domain/sale"
)

// The kinds of Promotion.
const (
	Percentage = "PERCENTAGE" // Value percent of the Sale's total, up to MaxDiscount
	Fixed      = "FIXED"      // Value baht off, never more than the total
)

var (
	ErrUnknownType       = errors.New("promotion type must be PERCENTAGE or FIXED")
	ErrValue             = errors.New("promotion value must be above 0 (a percentage at most 100), and its minimum and cap not negative")
	ErrBelowMinimum      = errors.New("order total below minimum purchase")
	ErrNoMatchingProduct = errors.New("no matching products for this promotion")
)

// Terms are what a Promotion gives and to which Sales.
type Terms struct {
	Type        string
	Value       float64
	MinPurchase float64
	MaxDiscount float64 // 0 is no cap
	ProductIds  []string
}

// Kind is the Type as stored: PERCENTAGE and FIXED in upper case.
func Kind(t string) string { return strings.ToUpper(strings.TrimSpace(t)) }

// Validate refuses Terms that would give nothing, or more than a Sale holds.
func (t Terms) Validate() error {
	switch Kind(t.Type) {
	case Percentage:
		if t.Value > 100 {
			return ErrValue
		}
	case Fixed:
	default:
		return ErrUnknownType
	}
	if t.Value <= 0 || t.MinPurchase < 0 || t.MaxDiscount < 0 {
		return ErrValue
	}
	return nil
}

// Discount is what the Terms take off a Sale of total holding productIds,
// rounded to the satang and never more than the total.
func (t Terms) Discount(total float64, productIds []string) (float64, error) {
	if err := t.Validate(); err != nil {
		return 0, err
	}
	if t.MinPurchase > 0 && total < t.MinPurchase {
		return 0, ErrBelowMinimum
	}
	if len(t.ProductIds) > 0 && len(productIds) > 0 && !holdsAny(productIds, t.ProductIds) {
		return 0, ErrNoMatchingProduct
	}
	discount := t.Value
	if Kind(t.Type) == Percentage {
		discount = total * t.Value / 100
		if t.MaxDiscount > 0 && discount > t.MaxDiscount {
			discount = t.MaxDiscount
		}
	}
	if discount > total {
		discount = total
	}
	return sale.Round(discount), nil
}

func holdsAny(held, wanted []string) bool {
	set := make(map[string]bool, len(wanted))
	for _, id := range wanted {
		set[id] = true
	}
	for _, id := range held {
		if set[id] {
			return true
		}
	}
	return false
}
