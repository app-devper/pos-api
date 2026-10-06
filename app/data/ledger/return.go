package ledger

import (
	"context"
	"math"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"pos/app/domain/stock"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Return puts goods from a completed Order back into the Stocks its Lines
// drew from, caps each refund at what the customer paid, and lets the goods
// serve any Line still waiting for that Unit.
func (l *Ledger) Return(ctx context.Context, req request.ProductReturn) (*entities.ProductReturn, error) {
	branch, err := primitive.ObjectIDFromHex(req.BranchId)
	if err != nil {
		return nil, reject("invalid branch id")
	}
	orderID, err := primitive.ObjectIDFromHex(req.OrderId)
	if err != nil {
		return nil, reject("invalid order id")
	}
	if len(req.Items) == 0 {
		return nil, reject("return must contain at least one Line")
	}
	seen := map[string]bool{}
	for _, line := range req.Items {
		if _, err := primitive.ObjectIDFromHex(line.OrderItemId); err != nil {
			return nil, reject("invalid Line id %q", line.OrderItemId)
		}
		if line.Quantity <= 0 || math.IsNaN(line.Refund) || math.IsInf(line.Refund, 0) {
			return nil, reject("invalid return quantity or refund")
		}
		if seen[line.OrderItemId] {
			return nil, reject("duplicate return Line %s", line.OrderItemId)
		}
		seen[line.OrderItemId] = true
	}
	result, err := l.run(ctx, func(b *book) (any, error) {
		return b.recordReturn(req, branch, orderID)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.ProductReturn), nil
}

func (b *book) recordReturn(req request.ProductReturn, branch, orderID primitive.ObjectID) (*entities.ProductReturn, error) {
	var order entities.Order
	err := b.col("orders").FindOne(b.ctx, bson.M{"_id": orderID, "branchId": branch}).Decode(&order)
	if err == mongo.ErrNoDocuments {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if !constant.IsConfirmedOrderStatus(order.Status) {
		return nil, reject("cannot return a cancelled Order")
	}
	// Touch the Order so cancelling it concurrently conflicts with this Return.
	if _, err := b.col("orders").UpdateOne(b.ctx, bson.M{"_id": orderID}, bson.M{"$inc": bson.M{"returnRevision": 1}}); err != nil {
		return nil, err
	}
	data := &entities.ProductReturn{Id: primitive.NewObjectID(), BranchId: branch, OrderId: orderID, CustomerCode: order.CustomerCode,
		Reason: req.Reason, Note: req.Note, CreatedBy: req.CreatedBy, CreatedDate: time.Now(), Items: []entities.ProductReturnItem{}}
	for _, line := range req.Items {
		id, _ := primitive.ObjectIDFromHex(line.OrderItemId)
		var item entities.OrderItem
		if err := b.col("order_items").FindOne(b.ctx, bson.M{"_id": id, "orderId": orderID}).Decode(&item); err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, ErrNotFound
			}
			return nil, err
		}
		if !constant.IsConfirmedOrderItemStatus(item.Status) {
			return nil, reject("cannot return a cancelled Line")
		}
		real := stock.RealQuantity(item.Stocks)
		if line.Quantity > real-item.ReturnedQty || line.Quantity > item.Quantity-item.ReturnedQty {
			return nil, reject("คืนได้สูงสุด %d", max(0, min(real, item.Quantity)-item.ReturnedQty))
		}
		maxRefund := paidPerUnit(&item) * float64(line.Quantity)
		if line.Refund < 0 || line.Refund > maxRefund+0.005 {
			return nil, reject("คืนเงินได้สูงสุด %.2f", maxRefund)
		}
		// Goods coming back from this Line serve other waiting Lines, never
		// this Line's own debt: that would give it Stock it never drew.
		b.notOwed = append(b.notOwed, item.Id)
		if _, err := b.col("order_items").UpdateOne(b.ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"returnedQty": line.Quantity}, "$set": bson.M{"updatedDate": time.Now()}}); err != nil {
			return nil, err
		}
		for _, back := range stock.AllocateReturn(item.Stocks, item.ReturnedQty, line.Quantity) {
			stockID, err := primitive.ObjectIDFromHex(back.StockId)
			if err != nil {
				return nil, err
			}
			st, err := b.stock(stockID, item.ProductId, branch)
			if err != nil {
				return nil, err
			}
			if st.UnitId != item.UnitId {
				return nil, reject("return Lot has a different Unit")
			}
			reason, by := req.Reason, req.CreatedBy
			if err := b.change(st, back.Quantity, func(qty int) request.ProductHistory {
				return request.ProductReturnHistory(st.ProductId.Hex(), "", qty, reason, 0, by)
			}); err != nil {
				return nil, err
			}
		}
		data.Items = append(data.Items, entities.ProductReturnItem{OrderItemId: id, ProductId: item.ProductId, Quantity: line.Quantity, Price: item.Price, Refund: line.Refund})
		data.TotalRefund += line.Refund
	}
	code, err := b.l.codes(b.ctx, constant.PRODUCT_RETURN, "RT-")
	if err != nil {
		return nil, err
	}
	data.ReturnNo = code
	if _, err := b.col("product_returns").InsertOne(b.ctx, data); err != nil {
		return nil, err
	}
	return data, nil
}

// paidPerUnit is what the customer paid for one unit of a Line: its price is
// the whole quantity before the per-unit discount.
func paidPerUnit(item *entities.OrderItem) float64 {
	if item.Quantity <= 0 {
		return 0
	}
	return math.Max(item.Price/float64(item.Quantity)-item.Discount, 0)
}
