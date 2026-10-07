package repositories

import (
	"strings"

	"pos/app/data/entities"
	"pos/app/domain/stock"
)

// A cancellation restores only the part a Return has not restored already.
// Synthetic Adjustment references have no Lot to restore, while sold-first
// allocations still reverse the Product's sold-first balance.
func cancellationStock(stocks []entities.OrderItemStock, returned int) []entities.OrderItemStock {
	result := make([]entities.OrderItemStock, 0, len(stocks))
	for _, s := range stocks {
		if strings.HasPrefix(s.StockId, stock.SyntheticPrefix) {
			continue
		}
		if s.StockId != "" {
			skip := minInt(returned, s.Quantity)
			s.Quantity -= skip
			returned -= skip
		}
		if s.Quantity > 0 {
			result = append(result, s)
		}
	}
	return result
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
