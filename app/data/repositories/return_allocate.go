package repositories

import (
	"strings"

	"pos/app/data/entities"
)

const syntheticStockRefPrefix = "ADJUST:"

func isRealLotReference(stockId string) bool {
	return stockId != "" && !strings.HasPrefix(stockId, syntheticStockRefPrefix)
}

func realLotQuantity(stocks []entities.OrderItemStock) int {
	total := 0
	for _, s := range stocks {
		if isRealLotReference(s.StockId) {
			total += s.Quantity
		}
	}
	return total
}

func allocateReturnAcrossRealLots(stocks []entities.OrderItemStock, alreadyReturned int, returnQty int) []entities.OrderItemStock {
	allocations := make([]entities.OrderItemStock, 0, len(stocks))
	skip := alreadyReturned
	remaining := returnQty
	for _, s := range stocks {
		if remaining <= 0 {
			break
		}
		if !isRealLotReference(s.StockId) {
			continue
		}
		qty := s.Quantity
		if skip > 0 {
			if skip >= qty {
				skip -= qty
				continue
			}
			qty -= skip
			skip = 0
		}
		take := qty
		if take > remaining {
			take = remaining
		}
		if take <= 0 {
			continue
		}
		allocations = append(allocations, entities.OrderItemStock{StockId: s.StockId, Quantity: take})
		remaining -= take
	}
	return allocations
}

// A cancellation restores only the part a Return has not restored already.
// Synthetic Adjustment references have no Lot to restore, while sold-first
// allocations still reverse the Product's sold-first balance.
func cancellationStock(stocks []entities.OrderItemStock, returned int) []entities.OrderItemStock {
	result := make([]entities.OrderItemStock, 0, len(stocks))
	for _, stock := range stocks {
		if strings.HasPrefix(stock.StockId, syntheticStockRefPrefix) {
			continue
		}
		if stock.StockId != "" {
			skip := minInt(returned, stock.Quantity)
			stock.Quantity -= skip
			returned -= skip
		}
		if stock.Quantity > 0 {
			result = append(result, stock)
		}
	}
	return result
}
