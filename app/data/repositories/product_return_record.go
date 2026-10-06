package repositories

import (
	"context"

	"pos/app/data/entities"
	"pos/app/domain/request"
)

// RecordProductReturn is recorded by the Stock ledger (ADR-0001).
func (entity *productReturnEntity) RecordProductReturn(req request.ProductReturn) (*entities.ProductReturn, error) {
	return entity.ledger.Return(context.Background(), req)
}
