package ledger

import (
	"context"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	transferPending  = "PENDING"
	transferApproved = "APPROVED"
	transferRejected = "REJECTED"
)

// RequestTransfer records a Transfer and reserves its quantity out of the
// source Stocks at once, so goods on their way to another branch cannot be
// sold meanwhile.
func (l *Ledger) RequestTransfer(ctx context.Context, form request.StockTransfer) (*entities.StockTransfer, error) {
	from, err := primitive.ObjectIDFromHex(form.FromBranchId)
	if err != nil {
		return nil, reject("invalid source branch id")
	}
	to, err := primitive.ObjectIDFromHex(form.ToBranchId)
	if err != nil {
		return nil, reject("invalid destination branch id")
	}
	if from == to {
		return nil, reject("สาขาปลายทางต้องไม่ใช่สาขาเดียวกับต้นทาง")
	}
	if len(form.Items) == 0 {
		return nil, reject("ใบโอนต้องมีอย่างน้อยหนึ่งรายการ")
	}
	items := make([]entities.StockTransferItem, len(form.Items))
	seen := map[string]bool{}
	for i, item := range form.Items {
		product, err := primitive.ObjectIDFromHex(item.ProductId)
		if err != nil {
			return nil, reject("invalid product id %q", item.ProductId)
		}
		// A line moves goods out of one named Stock; a line without one would
		// be recorded as transferred while nothing moves.
		if item.StockId == "" {
			return nil, reject("รายการโอนต้องระบุสต็อกที่โอนออก")
		}
		if seen[item.StockId] {
			return nil, reject("ใบโอนระบุสต็อกเดียวกันซ้ำ")
		}
		seen[item.StockId] = true
		if item.Quantity <= 0 {
			return nil, reject("จำนวนที่โอนต้องมากกว่า 0")
		}
		items[i] = entities.StockTransferItem{ProductId: product, StockId: item.StockId, Quantity: item.Quantity}
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		for _, item := range items {
			id, err := primitive.ObjectIDFromHex(item.StockId)
			if err != nil {
				return nil, reject("invalid stock id %q", item.StockId)
			}
			st, err := b.stock(id, item.ProductId, from)
			if err != nil {
				return nil, err
			}
			if st.Quantity < item.Quantity {
				return nil, reject("สต็อกไม่พอสำหรับโอนสินค้า %s (มี %d)", item.ProductId.Hex(), st.Quantity)
			}
			qty := item.Quantity
			if err := b.change(st, -qty, func(int) request.ProductHistory {
				return request.TransferStockHistory(st.ProductId.Hex(), constant.HistoryTypeTransferStockOut, form.Code, qty, form.CreatedBy)
			}); err != nil {
				return nil, err
			}
		}
		now := time.Now()
		data := &entities.StockTransfer{Id: primitive.NewObjectID(), FromBranchId: from, ToBranchId: to, Code: form.Code, Items: items,
			Note: form.Note, Status: transferPending, CreatedBy: form.CreatedBy, CreatedDate: now, UpdatedBy: form.CreatedBy, UpdatedDate: now}
		if _, err := b.col("stock_transfers").InsertOne(b.ctx, data); err != nil {
			return nil, err
		}
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.StockTransfer), nil
}

// ApproveTransfer opens a Stock at the destination for each reserved line —
// same Unit, lot, cost, price and expiry as its source — and lets it serve the
// destination's waiting Lines of that Unit (ADR-0002).
func (l *Ledger) ApproveTransfer(ctx context.Context, transferID, by string) (*entities.StockTransfer, error) {
	return l.settleTransfer(ctx, transferID, by, transferApproved, func(b *book, t *entities.StockTransfer, item entities.StockTransferItem, source *entities.ProductStock) error {
		sequence, err := b.nextSequence(source.ProductId, source.UnitId, t.ToBranchId)
		if err != nil {
			return err
		}
		qty := item.Quantity
		_, err = b.open(entities.ProductStock{Id: primitive.NewObjectID(), BranchId: t.ToBranchId, ProductId: source.ProductId, UnitId: source.UnitId,
			Sequence: sequence, LotNumber: source.LotNumber, CostPrice: source.CostPrice, Price: source.Price, Import: qty, Quantity: qty,
			ExpireDate: source.ExpireDate, ImportDate: time.Now()}, func(int) request.ProductHistory {
			return request.TransferStockHistory(source.ProductId.Hex(), constant.HistoryTypeTransferStockIn, t.Code, qty, by)
		})
		return err
	})
}

// RejectTransfer puts each reserved quantity back into its source Stock,
// where it serves the source's waiting Lines of that Unit (ADR-0002).
func (l *Ledger) RejectTransfer(ctx context.Context, transferID, by string) (*entities.StockTransfer, error) {
	return l.settleTransfer(ctx, transferID, by, transferRejected, func(b *book, t *entities.StockTransfer, item entities.StockTransferItem, source *entities.ProductStock) error {
		qty := item.Quantity
		return b.change(source, qty, func(int) request.ProductHistory {
			return request.TransferStockHistory(source.ProductId.Hex(), constant.HistoryTypeTransferStockBack, t.Code, qty, by)
		})
	})
}

// settleTransfer moves a PENDING Transfer to its final status, once, applying
// step to each reserved line in the same transaction.
func (l *Ledger) settleTransfer(ctx context.Context, transferID, by, status string, step func(*book, *entities.StockTransfer, entities.StockTransferItem, *entities.ProductStock) error) (*entities.StockTransfer, error) {
	id, err := primitive.ObjectIDFromHex(transferID)
	if err != nil {
		return nil, reject("invalid transfer id")
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		var t entities.StockTransfer
		if err := missing(b.col("stock_transfers").FindOne(b.ctx, bson.M{"_id": id}).Decode(&t), "transfer %s", transferID); err != nil {
			return nil, err
		}
		if t.Status != transferPending {
			return nil, reject("ใบโอนนี้ไม่อยู่ในสถานะรออนุมัติ")
		}
		for _, item := range t.Items {
			if item.StockId == "" {
				continue
			}
			stockID, err := primitive.ObjectIDFromHex(item.StockId)
			if err != nil {
				return nil, reject("invalid stock id %q", item.StockId)
			}
			source, err := b.stock(stockID, item.ProductId, t.FromBranchId)
			if err != nil {
				return nil, err
			}
			if err := step(b, &t, item, source); err != nil {
				return nil, err
			}
		}
		after := options.After
		var updated entities.StockTransfer
		err := b.col("stock_transfers").FindOneAndUpdate(b.ctx, bson.M{"_id": id, "status": transferPending},
			bson.M{"$set": bson.M{"status": status, "updatedBy": by, "updatedDate": time.Now()}},
			&options.FindOneAndUpdateOptions{ReturnDocument: &after}).Decode(&updated)
		if err == mongo.ErrNoDocuments {
			return nil, ErrConflict
		}
		if err != nil {
			return nil, err
		}
		return &updated, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.StockTransfer), nil
}
