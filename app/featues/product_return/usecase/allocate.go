package usecase

import (
	"strings"

	"pos/app/data/entities"
)

const syntheticStockRefPrefix = "ADJUST:"

func isSyntheticStockRef(stockId string) bool {
	return strings.HasPrefix(stockId, syntheticStockRefPrefix)
}

func realLotQuantity(stocks []entities.OrderItemStock) int {
	total := 0
	for _, s := range stocks {
		if !isSyntheticStockRef(s.StockId) {
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
		if isSyntheticStockRef(s.StockId) {
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
