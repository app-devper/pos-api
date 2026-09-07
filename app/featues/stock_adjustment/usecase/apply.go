package usecase

import (
	"fmt"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func IsValidAdjustmentReason(reason string) bool {
	for _, r := range constant.AdjustmentReasons() {
		if r == reason {
			return true
		}
	}
	return false
}

func ApplyAdjustment(
	stockAdjustmentEntity repositories.IStockAdjustment,
	productStock repositories.IProductStock,
	productEntity repositories.IProduct,
	orderEntity repositories.IOrder,
	sequenceEntity repositories.ISequence,
	req request.StockAdjustment,
) (*entities.StockAdjustment, error) {
	if req.Delta == 0 {
		return nil, fmt.Errorf("delta must be non-zero")
	}
	if !IsValidAdjustmentReason(req.Reason) {
		return nil, fmt.Errorf("invalid adjustment reason: %s", req.Reason)
	}

	stock, err := productStock.GetProductStockById(req.StockId)
	if err != nil {
		return nil, fmt.Errorf("failed to load stock %s: %w", req.StockId, err)
	}
	if stock == nil {
		return nil, fmt.Errorf("stock %s not found", req.StockId)
	}
	if req.BranchId != "" && stock.BranchId.Hex() != req.BranchId {
		return nil, fmt.Errorf("stock %s does not belong to branch %s", req.StockId, req.BranchId)
	}

	before := stock.Quantity
	var updatedStock *entities.ProductStock
	if req.Delta > 0 {
		updatedStock, err = productStock.AddProductStockQuantityById(req.StockId, req.Delta)
	} else {
		updatedStock, err = productStock.RemoveProductStockQuantityById(req.StockId, -req.Delta)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to apply adjustment to stock %s: %w", req.StockId, err)
	}
	if updatedStock == nil {
		return nil, fmt.Errorf("insufficient stock %s for adjustment", req.StockId)
	}

	sequence, err := sequenceEntity.NextSequence(constant.STOCK_ADJUSTMENT)
	if err != nil {
		return nil, fmt.Errorf("failed to generate adjustment code: %w", err)
	}
	if sequence == nil {
		return nil, fmt.Errorf("stock adjustment sequence not available")
	}
	code := "AJ-" + sequence.GenerateCode()

	unit, err := productEntity.GetProductUnitById(stock.UnitId.Hex())
	if err != nil || unit == nil {
		return nil, fmt.Errorf("failed to load unit for stock %s: %w", req.StockId, err)
	}
	balance := productStock.GetProductStockBalance(req.ProductId, unit.Id.Hex(), req.BranchId)
	history := request.AdjustStockHistory(req.ProductId, unit.Unit, req, balance)
	history.BranchId = req.BranchId
	if _, err := productStock.CreateProductHistory(history); err != nil {
		return nil, fmt.Errorf("failed to create stock history for %s: %w", req.ProductId, err)
	}

	productObjId, err := primitive.ObjectIDFromHex(req.ProductId)
	if err != nil {
		return nil, err
	}
	stockObjId, err := primitive.ObjectIDFromHex(req.StockId)
	if err != nil {
		return nil, err
	}
	branchObjId, err := primitive.ObjectIDFromHex(req.BranchId)
	if err != nil {
		return nil, err
	}

	adjustment, err := stockAdjustmentEntity.CreateStockAdjustment(repositories.StockAdjustmentInput{
		BranchId:  branchObjId,
		Code:      code,
		ProductId: productObjId,
		StockId:   stockObjId,
		Reason:    req.Reason,
		Note:      req.Note,
		Delta:     req.Delta,
		Before:    before,
		After:     updatedStock.Quantity,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to record stock adjustment for %s: %w", req.ProductId, err)
	}

	if req.Delta > 0 {
		reconcileOversoldFromAdjustment(orderEntity, req.ProductId, req.BranchId, req.Delta, "ADJUST:"+req.Reason)
	}

	return adjustment, nil
}

func reconcileOversoldFromAdjustment(orderEntity repositories.IOrder, productId string, branchId string, delta int, stockRef string) {
	if delta <= 0 {
		return
	}
	oversoldItems, err := orderEntity.GetOversoldOrderItemsByProductId(productId, branchId)
	if err != nil {
		logrus.WithError(err).WithField("productId", productId).Warn("failed to look up oversold order items during adjustment reconciliation")
		return
	}
	available := delta
	for _, item := range oversoldItems {
		if available <= 0 {
			break
		}
		if item.OversoldQty <= 0 {
			continue
		}
		drain := item.OversoldQty
		if available < drain {
			drain = available
		}
		if _, err := orderEntity.DrainOversoldQtyByOrderItemId(item.Id.Hex(), drain, stockRef); err != nil {
			logrus.WithError(err).WithField("orderItemId", item.Id.Hex()).Warn("failed to record oversold reconciliation from adjustment")
			continue
		}
		available -= drain
	}
}
