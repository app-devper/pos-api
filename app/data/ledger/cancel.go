package ledger

import (
	"context"
	"errors"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/app/domain/stock"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// CancelOrder cancels a whole Order: its Lines, payments and the Order itself.
// Every standing Line puts back what it drew — settlement draws included —
// and that quantity serves other Lines waiting for the Unit (ADR-0002). A
// Line's own unsettled debt simply ends.
func (l *Ledger) CancelOrder(ctx context.Context, orderID, branchID, by, reason string) error {
	id, branch, err := ids(orderID, branchID)
	if err != nil {
		return err
	}
	_, err = l.run(ctx, func(b *book) (any, error) {
		var order entities.Order
		if err := missing(b.col("orders").FindOne(b.ctx, bson.M{"_id": id, "branchId": branch}).Decode(&order), "order %s in this branch", orderID); err != nil {
			return nil, err
		}
		if !constant.IsConfirmedOrderStatus(order.Status) {
			return nil, reject("บิลนี้ถูกยกเลิกไปแล้ว")
		}
		var lines []entities.OrderItem
		if err := find(b, "order_items", bson.M{"orderId": id, "$or": confirmedLineStatuses()}, &lines); err != nil {
			return nil, err
		}
		cancelled := cancelledBy(by, reason)
		for _, col := range []string{"order_items", "payments"} {
			if _, err := b.col(col).UpdateMany(b.ctx, bson.M{"orderId": id}, cancelled); err != nil {
				return nil, err
			}
		}
		res, err := b.col("orders").UpdateOne(b.ctx, bson.M{"_id": id, "status": bson.M{"$in": constant.ConfirmedOrderStatuses()}}, cancelled)
		if err != nil {
			return nil, err
		}
		if res.MatchedCount != 1 {
			return nil, ErrConflict
		}
		for i := range lines {
			if err := b.putBack(&lines[i], by); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

// CancelLine cancels one Line of an Order, puts back what it drew (as
// CancelOrder does) and recomputes the Order's money from the Lines left.
func (l *Ledger) CancelLine(ctx context.Context, lineID, branchID, by, reason string) error {
	id, branch, err := ids(lineID, branchID)
	if err != nil {
		return err
	}
	_, err = l.run(ctx, func(b *book) (any, error) {
		var line entities.OrderItem
		if err := missing(b.col("order_items").FindOne(b.ctx, bson.M{"_id": id, "branchId": branch}).Decode(&line), "Line %s in this branch", lineID); err != nil {
			return nil, err
		}
		if !constant.IsConfirmedOrderItemStatus(line.Status) {
			return nil, reject("รายการนี้ถูกยกเลิกไปแล้ว")
		}
		res, err := b.col("order_items").UpdateOne(b.ctx, bson.M{"_id": id, "$or": confirmedLineStatuses()}, cancelledBy(by, reason))
		if err != nil {
			return nil, err
		}
		if res.MatchedCount != 1 {
			return nil, ErrConflict
		}
		if err := b.putBack(&line, by); err != nil {
			return nil, err
		}
		return nil, b.recomputeOrder(line.OrderId)
	})
	return err
}

// putBack returns what a cancelled Line drew to the Stocks it came from, and
// Sold first parts to Sold first, under one history row for the Line. A Stock
// or Product deleted by hand since the sale has nothing to take quantity back;
// that part is skipped rather than blocking the cancel.
func (b *book) putBack(line *entities.OrderItem, by string) error {
	detail := entities.OrderItemProductDetail{Quantity: max(0, line.Quantity-line.ReturnedQty), Price: line.Price, CostPrice: line.CostPrice}
	r := b.rowFor(line.Id, &entities.ProductStock{BranchId: line.BranchId, ProductId: line.ProductId, UnitId: line.UnitId}, func(int) request.ProductHistory {
		return request.RemoveOrderItemProductHistory(line.ProductId.Hex(), "", &detail, 0, by)
	})
	for _, part := range stock.CancellationParts(line.Stocks, line.ReturnedQty) {
		if part.StockId == "" {
			var rejected *Rejected
			if err := b.soldFirst(line.ProductId, part.Quantity); err != nil && !errors.As(err, &rejected) {
				return err
			}
			continue
		}
		stockID, err := primitive.ObjectIDFromHex(part.StockId)
		if err != nil {
			continue // not a Stock id: nothing to put back into
		}
		var st entities.ProductStock
		err = b.col("product_stocks").FindOne(b.ctx, bson.M{"_id": stockID}).Decode(&st)
		if errors.Is(err, mongo.ErrNoDocuments) {
			continue
		}
		if err != nil {
			return err
		}
		if err := b.move(&st, part.Quantity, r); err != nil {
			return err
		}
	}
	return nil
}

// recomputeOrder sets an Order's money from its standing Lines, summed the
// way a Sale sums them.
func (b *book) recomputeOrder(order primitive.ObjectID) error {
	var lines []entities.OrderItem
	if err := find(b, "order_items", bson.M{"orderId": order, "$or": confirmedLineStatuses()}, &lines); err != nil {
		return err
	}
	var money lineMoney
	for _, line := range lines {
		qty := float64(line.Quantity)
		money.add(roundMoney(line.Price-line.Discount*qty), line.CostPrice, line.Discount*qty)
	}
	total, cost, discount := money.rounded()
	_, err := b.col("orders").UpdateOne(b.ctx, bson.M{"_id": order}, bson.M{"$set": bson.M{
		"total": total, "totalCost": cost, "discount": discount, "updatedDate": time.Now(),
	}})
	return err
}

func cancelledBy(by, reason string) bson.M {
	return bson.M{"$set": bson.M{"status": constant.CANCELLED, "cancelReason": reason, "updatedBy": by, "updatedDate": time.Now()}}
}

func ids(docID, branchID string) (primitive.ObjectID, primitive.ObjectID, error) {
	id, err := primitive.ObjectIDFromHex(docID)
	if err != nil {
		return id, id, reject("invalid id %q", docID)
	}
	branch, err := primitive.ObjectIDFromHex(branchID)
	if err != nil {
		return id, branch, reject("invalid branch id")
	}
	return id, branch, nil
}
