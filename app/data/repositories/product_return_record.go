package repositories

import (
	"context"
	"fmt"
	"math"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecordProductReturn validates and restores the original Lots, advances the
// returned quantities, and writes the Return and history in one transaction.
func (entity *productReturnEntity) RecordProductReturn(req request.ProductReturn) (*entities.ProductReturn, error) {
	branch, err := primitive.ObjectIDFromHex(req.BranchId)
	if err != nil {
		return nil, err
	}
	orderID, err := primitive.ObjectIDFromHex(req.OrderId)
	if err != nil {
		return nil, err
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("return must contain at least one Line")
	}
	seen := map[primitive.ObjectID]bool{}
	for _, line := range req.Items {
		if _, err := primitive.ObjectIDFromHex(line.OrderItemId); err != nil {
			return nil, err
		}
		if line.Quantity <= 0 || math.IsNaN(line.Refund) || math.IsInf(line.Refund, 0) {
			return nil, fmt.Errorf("invalid return quantity or refund")
		}
		lineID, _ := primitive.ObjectIDFromHex(line.OrderItemId)
		if seen[lineID] {
			return nil, fmt.Errorf("duplicate return Line %s", line.OrderItemId)
		}
		seen[lineID] = true
	}
	result, err := entity.recorder.transaction(func(ctx context.Context) (interface{}, error) {
		return entity.recordProductReturn(ctx, req, branch, orderID)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.ProductReturn), nil
}

func (entity *productReturnEntity) recordProductReturn(ctx context.Context, req request.ProductReturn, branch, orderID primitive.ObjectID) (*entities.ProductReturn, error) {
	r := entity.recorder
	var order entities.Order
	if err := r.db.Collection("orders").FindOne(ctx, bson.M{"_id": orderID, "branchId": branch}).Decode(&order); err != nil {
		return nil, fmt.Errorf("order not found in this branch: %w", err)
	}
	if !constant.IsConfirmedOrderStatus(order.Status) {
		return nil, fmt.Errorf("cannot return a cancelled Order")
	}
	// Touch the Order so cancelling it concurrently conflicts with this command.
	if _, err := r.db.Collection("orders").UpdateOne(ctx, bson.M{"_id": orderID}, bson.M{"$inc": bson.M{"returnRevision": 1}}); err != nil {
		return nil, err
	}
	data := &entities.ProductReturn{Id: primitive.NewObjectID(), BranchId: branch, OrderId: orderID, CustomerCode: order.CustomerCode, Reason: req.Reason, Note: req.Note, CreatedBy: req.CreatedBy, CreatedDate: time.Now(), Items: []entities.ProductReturnItem{}}
	for _, line := range req.Items {
		id, _ := primitive.ObjectIDFromHex(line.OrderItemId)
		var item entities.OrderItem
		if err := r.db.Collection("order_items").FindOne(ctx, bson.M{"_id": id, "orderId": orderID}).Decode(&item); err != nil {
			return nil, fmt.Errorf("Line not found on Order: %w", err)
		}
		if !constant.IsConfirmedOrderItemStatus(item.Status) {
			return nil, fmt.Errorf("cannot return a cancelled Line")
		}
		realQty := realLotQuantity(item.Stocks)
		if line.Quantity > realQty-item.ReturnedQty || line.Quantity > item.Quantity-item.ReturnedQty {
			return nil, fmt.Errorf("คืนได้สูงสุด %d", maxInt(0, minInt(realQty, item.Quantity)-item.ReturnedQty))
		}
		maxRefund := paidPerUnit(&item) * float64(line.Quantity)
		if line.Refund < 0 || line.Refund > maxRefund+0.005 {
			return nil, fmt.Errorf("คืนเงินได้สูงสุด %.2f", maxRefund)
		}
		allocations := allocateReturnAcrossRealLots(item.Stocks, item.ReturnedQty, line.Quantity)
		for _, allocation := range allocations {
			stockID, err := primitive.ObjectIDFromHex(allocation.StockId)
			if err != nil {
				return nil, err
			}
			stock, err := r.stock(ctx, stockID, item.ProductId, branch)
			if err != nil {
				return nil, err
			}
			if stock.UnitId != item.UnitId {
				return nil, fmt.Errorf("return Lot has a different Unit")
			}
			if _, err := r.db.Collection("product_stocks").UpdateOne(ctx, bson.M{"_id": stockID}, bson.M{"$inc": bson.M{"quantity": allocation.Quantity}}); err != nil {
				return nil, err
			}
		}
		if _, err := r.db.Collection("order_items").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"returnedQty": line.Quantity}, "$set": bson.M{"updatedDate": time.Now()}}); err != nil {
			return nil, err
		}
		h := request.ProductReturnHistory(item.ProductId.Hex(), "", line.Quantity, req.Reason, 0, req.CreatedBy)
		if err := r.history(ctx, &entities.ProductStock{ProductId: item.ProductId, UnitId: item.UnitId, BranchId: branch}, h); err != nil {
			return nil, err
		}
		data.Items = append(data.Items, entities.ProductReturnItem{OrderItemId: id, ProductId: item.ProductId, Quantity: line.Quantity, Price: item.Price, Refund: line.Refund})
		data.TotalRefund += line.Refund
	}
	code, err := r.nextCode(ctx, constant.PRODUCT_RETURN, "RT-")
	if err != nil {
		return nil, err
	}
	data.ReturnNo = code
	if _, err := entity.productReturnRepo.InsertOne(ctx, data); err != nil {
		return nil, err
	}
	return data, nil
}

func paidPerUnit(item *entities.OrderItem) float64 {
	if item.Quantity <= 0 {
		return 0
	}
	return math.Max(item.Price/float64(item.Quantity)-item.Discount, 0)
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
