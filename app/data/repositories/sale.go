package repositories

import (
	"context"

	"pos/app/data/entities"
	"pos/app/data/ledger"
	"pos/app/domain/request"

	"github.com/sirupsen/logrus"
)

// ErrSaleConflict: the Sale's id was already recorded for a different Sale.
var ErrSaleConflict = ledger.ErrSaleConflict

// SaleRejected is a Sale that cannot be recorded as it stands.
type SaleRejected = ledger.Rejected

// RecordedSale is the Order a Sale was recorded as, with the current state
// of the Stocks it drew from.
type RecordedSale struct {
	Order  *entities.Order
	Stocks []entities.ProductStock
}

// RecordSale is recorded by the Stock ledger (ADR-0001): pricing, drawing,
// the Order code and the repeat check all happen in one transaction.
func (entity *orderEntity) RecordSale(form request.Sale) (*RecordedSale, error) {
	logrus.Info("RecordSale")
	sold, err := entity.ledger.Sell(context.Background(), form)
	if err != nil {
		return nil, err
	}
	return &RecordedSale{Order: sold.Order, Stocks: sold.Stocks}, nil
}
