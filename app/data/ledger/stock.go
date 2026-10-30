package ledger

import (
	"context"
	"errors"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateStock adds a Stock by hand — the way a Unit other than the main one
// gets Stock (ADR-0002 note) — and lets it serve that Unit's waiting Lines.
func (l *Ledger) CreateStock(ctx context.Context, req request.ProductStock) (*entities.ProductStock, error) {
	branch, err := primitive.ObjectIDFromHex(req.BranchId)
	if err != nil {
		return nil, reject("invalid branch id")
	}
	product, err := primitive.ObjectIDFromHex(req.ProductId)
	if err != nil {
		return nil, reject("invalid product id")
	}
	unit, err := primitive.ObjectIDFromHex(req.UnitId)
	if err != nil {
		return nil, reject("invalid unit id")
	}
	if req.Quantity < 0 {
		return nil, reject("จำนวนสต็อกต้องไม่ติดลบ")
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		err := b.col("product_units").FindOne(b.ctx, bson.M{"_id": unit, "productId": product}).Err()
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, reject("unit %s is not a unit of product %s", req.UnitId, req.ProductId)
		}
		if err != nil {
			return nil, err
		}
		sequence, err := b.nextSequence(product, unit, branch)
		if err != nil {
			return nil, err
		}
		return b.open(entities.ProductStock{Id: primitive.NewObjectID(), BranchId: branch, ProductId: product, UnitId: unit,
			ReceiveCode: req.ReceiveCode, Sequence: sequence, LotNumber: req.LotNumber, CostPrice: req.CostPrice, Price: req.Price,
			Import: req.Quantity, Quantity: req.Quantity, ExpireDate: req.ExpireDate.Time, ImportDate: req.ImportDate.Time},
			func(int) request.ProductHistory { return request.AddProductStockHistory(req.ProductId, "", req, 0) })
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.ProductStock), nil
}

// DeleteStock removes a Stock that holds nothing. A Stock still holding goods
// is corrected with an Adjustment or a Count instead.
func (l *Ledger) DeleteStock(ctx context.Context, stockID, branchID, by string) (*entities.ProductStock, error) {
	id, branch, err := ids(stockID, branchID)
	if err != nil {
		return nil, err
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		var st entities.ProductStock
		if err := missing(b.col("product_stocks").FindOne(b.ctx, bson.M{"_id": id, "branchId": branch}).Decode(&st), "stock %s in this branch", stockID); err != nil {
			return nil, err
		}
		if st.Quantity > 0 {
			return nil, reject("ลบสต็อกที่ยังมีสินค้าเหลือไม่ได้")
		}
		// A pending Transfer reserved out of this Stock and puts it back here
		// on reject, so the Stock must outlive the Transfer.
		if n, err := b.col("stock_transfers").CountDocuments(b.ctx, bson.M{"status": transferPending, "items.stockId": stockID}); err != nil {
			return nil, err
		} else if n > 0 {
			return nil, reject("สต็อกนี้ถูกจองไว้ในใบโอนที่รออนุมัติ ลบไม่ได้")
		}
		res, err := b.col("product_stocks").DeleteOne(b.ctx, bson.M{"_id": id, "quantity": st.Quantity})
		if err != nil {
			return nil, err
		}
		if res.DeletedCount != 1 {
			return nil, ErrConflict
		}
		b.rowFor(st.Id, &st, func(int) request.ProductHistory {
			return request.RemoveProductStockHistory(st.ProductId.Hex(), "", &st, 0, by)
		})
		return &st, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.ProductStock), nil
}

// SetQuantity records what one Stock actually holds. It is a one-Line Count:
// an Adjustment of the difference, with history, and a rise settles that
// Unit's waiting Lines (ADR-0001).
func (l *Ledger) SetQuantity(ctx context.Context, stockID, branchID string, quantity int, by string) (*entities.ProductStock, error) {
	id, branch, err := ids(stockID, branchID)
	if err != nil {
		return nil, err
	}
	if quantity < 0 {
		return nil, reject("จำนวนสต็อกต้องไม่ติดลบ")
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		var st entities.ProductStock
		if err := missing(b.col("product_stocks").FindOne(b.ctx, bson.M{"_id": id, "branchId": branch}).Decode(&st), "stock %s in this branch", stockID); err != nil {
			return nil, err
		}
		if delta := quantity - st.Quantity; delta != 0 {
			if _, err := b.adjust(&st, request.StockAdjustment{ProductId: st.ProductId.Hex(), StockId: stockID, BranchId: branchID,
				Reason: constant.AdjustmentReasonCount, Delta: delta, CreatedBy: by}); err != nil {
				return nil, err
			}
		}
		return &st, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.ProductStock), nil
}
