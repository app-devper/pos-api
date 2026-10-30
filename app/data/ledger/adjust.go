package ledger

import (
	"context"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Adjust corrects one Stock's quantity by a reasoned delta. A rise settles
// that Unit's waiting Lines from the Stock.
func (l *Ledger) Adjust(ctx context.Context, req request.StockAdjustment) (*entities.StockAdjustment, error) {
	if err := validAdjustment(req); err != nil {
		return nil, err
	}
	branch, _ := primitive.ObjectIDFromHex(req.BranchId)
	product, _ := primitive.ObjectIDFromHex(req.ProductId)
	id, _ := primitive.ObjectIDFromHex(req.StockId)
	result, err := l.run(ctx, func(b *book) (any, error) {
		st, err := b.stock(id, product, branch)
		if err != nil {
			return nil, err
		}
		return b.adjust(st, req)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.StockAdjustment), nil
}

// Count records what the shelf holds. Every Stock whose count differs gets an
// Adjustment; all of it commits together or not at all.
func (l *Ledger) Count(ctx context.Context, req request.StockCount) (*entities.StockCount, error) {
	branch, err := primitive.ObjectIDFromHex(req.BranchId)
	if err != nil {
		return nil, reject("invalid branch id")
	}
	if len(req.Items) == 0 {
		return nil, reject("count must contain at least one Line")
	}
	seen := map[string]bool{}
	for _, line := range req.Items {
		for _, id := range []string{line.ProductId, line.StockId} {
			if _, err := primitive.ObjectIDFromHex(id); err != nil {
				return nil, reject("invalid id %q", id)
			}
		}
		if line.Counted < 0 {
			return nil, reject("counted quantity must not be negative")
		}
		if seen[line.StockId] {
			return nil, reject("duplicate Stock in Count")
		}
		seen[line.StockId] = true
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		code, err := l.codes(b.ctx, constant.STOCK_COUNT, "SC-")
		if err != nil {
			return nil, err
		}
		count := &entities.StockCount{Id: primitive.NewObjectID(), BranchId: branch, CountNo: code, Note: req.Note, Items: []entities.StockCountItem{}, CreatedBy: req.CreatedBy, CreatedDate: time.Now()}
		for _, line := range req.Items {
			product, _ := primitive.ObjectIDFromHex(line.ProductId)
			id, _ := primitive.ObjectIDFromHex(line.StockId)
			st, err := b.stock(id, product, branch)
			if err != nil {
				return nil, err
			}
			// Even an unchanged Line must name a real Unit of its Product.
			var unit entities.ProductUnit
			if err := missing(b.col("product_units").FindOne(b.ctx, map[string]any{"_id": st.UnitId, "productId": product}).Decode(&unit), "Unit of stock %s", id.Hex()); err != nil {
				return nil, err
			}
			system := st.Quantity
			delta := line.Counted - system
			if delta != 0 {
				if _, err := b.adjust(st, request.StockAdjustment{ProductId: line.ProductId, StockId: line.StockId, Reason: constant.AdjustmentReasonCount, Note: code + " " + req.Note, Delta: delta, BranchId: req.BranchId, CreatedBy: req.CreatedBy}); err != nil {
					return nil, err
				}
			}
			count.Items = append(count.Items, entities.StockCountItem{ProductId: product, StockId: id, SystemQuantity: system, CountedQuantity: line.Counted, Delta: delta})
		}
		if _, err := b.col("stock_counts").InsertOne(b.ctx, count); err != nil {
			return nil, err
		}
		return count, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.StockCount), nil
}

// adjust records one Adjustment. Before/After describe the correction itself;
// any settlement it triggers is recorded on the Lines it served.
func (b *book) adjust(st *entities.ProductStock, req request.StockAdjustment) (*entities.StockAdjustment, error) {
	before := st.Quantity
	if err := b.change(st, req.Delta, func(qty int) request.ProductHistory {
		r := req
		r.Delta = qty
		return request.AdjustStockHistory(st.ProductId.Hex(), "", r, 0)
	}); err != nil {
		return nil, err
	}
	code, err := b.l.codes(b.ctx, constant.STOCK_ADJUSTMENT, "AJ-")
	if err != nil {
		return nil, err
	}
	data := &entities.StockAdjustment{Id: primitive.NewObjectID(), BranchId: st.BranchId, Code: code, ProductId: st.ProductId, StockId: st.Id,
		Reason: req.Reason, Note: req.Note, Delta: req.Delta, Before: before, After: before + req.Delta, CreatedBy: req.CreatedBy, CreatedDate: time.Now()}
	if _, err := b.col("stock_adjustments").InsertOne(b.ctx, data); err != nil {
		return nil, err
	}
	return data, nil
}

func validAdjustment(req request.StockAdjustment) error {
	if req.Delta == 0 {
		return reject("delta must be non-zero")
	}
	valid := false
	for _, reason := range constant.AdjustmentReasons() {
		if req.Reason == reason {
			valid = true
			break
		}
	}
	if !valid {
		return reject("invalid adjustment reason: %s", req.Reason)
	}
	for _, id := range []string{req.BranchId, req.ProductId, req.StockId} {
		if _, err := primitive.ObjectIDFromHex(id); err != nil {
			return reject("invalid id %q", id)
		}
	}
	return nil
}
