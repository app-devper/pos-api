package repositories

import (
	"context"

	"pos/app/data/entities"
	"pos/app/domain/request"
)

// ApplyStockAdjustment is recorded by the Stock ledger (ADR-0001).
func (entity *stockAdjustmentEntity) ApplyStockAdjustment(req request.StockAdjustment) (*entities.StockAdjustment, error) {
	return entity.ledger.Adjust(context.Background(), req)
}
