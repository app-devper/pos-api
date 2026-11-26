package ledger

import (
	"context"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// deleteProductNote is the Adjustment note that says why a Stock went to zero.
const deleteProductNote = "ลบสินค้า"

// DeleteProduct marks a Product deleted and brings every Stock it still holds,
// in every branch, to zero with an Adjustment and a history row each, so the
// stock card says why the goods are gone. All of it commits together.
func (l *Ledger) DeleteProduct(ctx context.Context, productID, by string) (*entities.Product, error) {
	id, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, reject("invalid product id")
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		b.unitMayBeGone = true
		if err := b.refuseIfInPendingTransfer(id); err != nil {
			return nil, err
		}
		product, err := b.markDeleted(id, by)
		if err != nil {
			return nil, err
		}
		var stocks []entities.ProductStock
		if err := find(b, "product_stocks", bson.M{"productId": id, "quantity": bson.M{"$ne": 0}}, &stocks); err != nil {
			return nil, err
		}
		for i := range stocks {
			if err := b.zero(&stocks[i], by); err != nil {
				return nil, err
			}
		}
		return product, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.Product), nil
}

// refuseIfInPendingTransfer rejects deleting a Product a pending Transfer
// holds: its reservation has left the source Stock already, and a reject would
// put it back into, or an approve open, Stock of a deleted Product.
func (b *book) refuseIfInPendingTransfer(id primitive.ObjectID) error {
	n, err := b.col("stock_transfers").CountDocuments(b.ctx, bson.M{"status": transferPending, "items.productId": id})
	if err != nil {
		return err
	}
	if n > 0 {
		return reject("สินค้านี้อยู่ในใบโอนที่รออนุมัติ ลบไม่ได้")
	}
	return nil
}

// markDeleted sets deletedDate on a Product not deleted yet and returns it.
func (b *book) markDeleted(id primitive.ObjectID, by string) (*entities.Product, error) {
	now := time.Now()
	var product entities.Product
	err := b.col("products").FindOneAndUpdate(b.ctx,
		bson.M{"_id": id, "deletedDate": bson.M{"$exists": false}},
		bson.M{"$set": bson.M{"deletedDate": now, "updatedBy": by, "updatedDate": now}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&product)
	if err := missing(err, "product %s", id.Hex()); err != nil {
		return nil, err
	}
	return &product, nil
}

// zero brings one Stock of a deleted Product to zero through an Adjustment.
func (b *book) zero(st *entities.ProductStock, by string) error {
	_, err := b.adjustWith(st, request.StockAdjustment{ProductId: st.ProductId.Hex(), StockId: st.Id.Hex(), BranchId: st.BranchId.Hex(),
		Reason: constant.AdjustmentReasonOther, Note: deleteProductNote, Delta: -st.Quantity, CreatedBy: by}, request.DeleteProductStockHistory)
	return err
}
