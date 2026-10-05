package repositories

import (
	"context"
	"fmt"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (entity *stockAdjustmentEntity) ApplyStockAdjustment(req request.StockAdjustment) (*entities.StockAdjustment, error) {
	if err := validateAdjustment(req); err != nil {
		return nil, err
	}
	branch, _ := primitive.ObjectIDFromHex(req.BranchId)
	product, _ := primitive.ObjectIDFromHex(req.ProductId)
	id, _ := primitive.ObjectIDFromHex(req.StockId)
	result, err := entity.recorder.transaction(func(ctx context.Context) (interface{}, error) {
		stock, err := entity.recorder.stock(ctx, id, product, branch)
		if err != nil {
			return nil, err
		}
		return entity.recorder.applyAdjustment(ctx, stock, req)
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.StockAdjustment), nil
}

func validateAdjustment(req request.StockAdjustment) error {
	if req.Delta == 0 {
		return fmt.Errorf("delta must be non-zero")
	}
	valid := false
	for _, reason := range constant.AdjustmentReasons() {
		if req.Reason == reason {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid adjustment reason: %s", req.Reason)
	}
	for _, id := range []string{req.BranchId, req.ProductId, req.StockId} {
		if _, err := primitive.ObjectIDFromHex(id); err != nil {
			return err
		}
	}
	return nil
}

// Count and manual Adjustment use the same policy inside their own transaction.
func (r *stockRecorder) applyAdjustment(ctx context.Context, stock *entities.ProductStock, req request.StockAdjustment) (*entities.StockAdjustment, error) {
	after := stock.Quantity + req.Delta
	if after < 0 || (req.Delta > 0 && after < stock.Quantity) || (req.Delta < 0 && after > stock.Quantity) {
		return nil, fmt.Errorf("insufficient stock or quantity overflow")
	}
	result, err := r.db.Collection("product_stocks").UpdateOne(ctx, bson.M{"_id": stock.Id, "quantity": stock.Quantity}, bson.M{"$inc": bson.M{"quantity": req.Delta}})
	if err != nil {
		return nil, err
	}
	if result.MatchedCount != 1 {
		return nil, fmt.Errorf("stock changed during adjustment")
	}
	code, err := r.nextCode(ctx, constant.STOCK_ADJUSTMENT, "AJ-")
	if err != nil {
		return nil, err
	}
	if err := r.history(ctx, stock, request.AdjustStockHistory(stock.ProductId.Hex(), "", req, 0)); err != nil {
		return nil, err
	}
	data := &entities.StockAdjustment{Id: primitive.NewObjectID(), BranchId: stock.BranchId, Code: code, ProductId: stock.ProductId, StockId: stock.Id, Reason: req.Reason, Note: req.Note, Delta: req.Delta, Before: stock.Quantity, After: after, CreatedBy: req.CreatedBy, CreatedDate: time.Now()}
	if _, err := r.db.Collection("stock_adjustments").InsertOne(ctx, data); err != nil {
		return nil, err
	}
	if req.Delta > 0 {
		if err := r.reconcileAdjustment(ctx, stock.ProductId, stock.BranchId, req.Delta, "ADJUST:"+req.Reason); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (r *stockRecorder) reconcileAdjustment(ctx context.Context, product, branch primitive.ObjectID, delta int, ref string) error {
	cursor, err := r.db.Collection("order_items").Find(ctx, bson.M{"productId": product, "branchId": branch, "oversoldQty": bson.M{"$gt": 0}, "$or": confirmedOrderItemStatusMatchClauses()}, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	var items []entities.OrderItem
	if err := cursor.All(ctx, &items); err != nil {
		return err
	}
	for _, item := range items {
		if delta <= 0 {
			break
		}
		drain := minInt(delta, item.OversoldQty)
		result, err := r.db.Collection("order_items").UpdateOne(ctx, bson.M{"_id": item.Id, "oversoldQty": bson.M{"$gte": drain}}, mongo.Pipeline{{{Key: "$set", Value: bson.M{"oversoldQty": bson.M{"$subtract": bson.A{"$oversoldQty", drain}}, "stocks": bson.M{"$concatArrays": bson.A{bson.M{"$ifNull": bson.A{"$stocks", bson.A{}}}, bson.A{entities.OrderItemStock{Quantity: drain, StockId: ref}}}}, "updatedDate": time.Now()}}}})
		if err != nil {
			return err
		}
		if result.MatchedCount != 1 {
			return fmt.Errorf("oversold quantity changed during adjustment")
		}
		delta -= drain
	}
	return nil
}
