package repositories

import (
	"context"

	"pos/app/data/entities"
	"pos/app/domain/request"
)

// RecordStockCount is recorded by the Stock ledger (ADR-0001).
func (entity *stockCountEntity) RecordStockCount(req request.StockCount) (*entities.StockCount, error) {
	return entity.ledger.Count(context.Background(), req)
}
