package repositories

import (
	"context"
	"fmt"
	"time"

	"pos/app/data/entities"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RecordStockCount computes deltas from one snapshot and commits the entire
// Count, its Adjustments, Stock, history and reconciliation together.
func (entity *stockCountEntity) RecordStockCount(req request.StockCount) (*entities.StockCount, error) {
	branch, err := primitive.ObjectIDFromHex(req.BranchId)
	if err != nil {
		return nil, err
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("count must contain at least one Line")
	}
	seen := map[primitive.ObjectID]bool{}
	for _, line := range req.Items {
		for _, id := range []string{line.ProductId, line.StockId} {
			if _, err := primitive.ObjectIDFromHex(id); err != nil {
				return nil, err
			}
		}
		if line.Counted < 0 {
			return nil, fmt.Errorf("counted quantity must not be negative")
		}
		lineID, _ := primitive.ObjectIDFromHex(line.StockId)
		if seen[lineID] {
			return nil, fmt.Errorf("duplicate Stock in Count")
		}
		seen[lineID] = true
	}
	result, err := entity.recorder.transaction(func(ctx context.Context) (interface{}, error) {
		r := entity.recorder
		code, err := r.nextCode(ctx, constant.STOCK_COUNT, "SC-")
		if err != nil {
			return nil, err
		}
		data := &entities.StockCount{Id: primitive.NewObjectID(), BranchId: branch, CountNo: code, Note: req.Note, Items: []entities.StockCountItem{}, CreatedBy: req.CreatedBy, CreatedDate: time.Now()}
		for _, line := range req.Items {
			product, _ := primitive.ObjectIDFromHex(line.ProductId)
			id, _ := primitive.ObjectIDFromHex(line.StockId)
			stock, err := r.stock(ctx, id, product, branch)
			if err != nil {
				return nil, err
			}
			// Even unchanged Lines must reference a real Unit of the Product.
			var unit entities.ProductUnit
			if err := r.db.Collection("product_units").FindOne(ctx, map[string]interface{}{"_id": stock.UnitId, "productId": product}).Decode(&unit); err != nil {
				return nil, err
			}
			delta := line.Counted - stock.Quantity
			if delta != 0 {
				_, err := r.applyAdjustment(ctx, stock, request.StockAdjustment{ProductId: line.ProductId, StockId: line.StockId, Reason: constant.AdjustmentReasonCount, Note: code + " " + req.Note, Delta: delta, BranchId: req.BranchId, CreatedBy: req.CreatedBy})
				if err != nil {
					return nil, err
				}
			}
			data.Items = append(data.Items, entities.StockCountItem{ProductId: product, StockId: id, SystemQuantity: stock.Quantity, CountedQuantity: line.Counted, Delta: delta})
		}
		if _, err := entity.stockCountRepo.InsertOne(ctx, data); err != nil {
			return nil, err
		}
		return data, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*entities.StockCount), nil
}
