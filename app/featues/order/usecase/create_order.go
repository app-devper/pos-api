package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type stockAdjustment struct {
	stockId  string
	quantity int
}

type soldFirstAdjustment struct {
	productId string
	quantity  int
}

func CreateOrder(
	orderEntity repositories.IOrder,
	productEntity repositories.IProduct,
	productStock repositories.IProductStock,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Order{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		req.CreatedBy = userId
		req.BranchId = utils.GetBranchId(ctx)

		sequence, err := sequenceEntity.NextSequence(constant.ORDER)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}
		if sequence == nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, "order sequence not available")
			return
		}
		req.Code = sequence.GenerateCode()

		result, err := orderEntity.CreateOrder(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		orderCreated := true
		var updatedStocks []entities.ProductStock
		var stockAdjustments []stockAdjustment
		var soldFirstAdjustments []soldFirstAdjustment
		var createdHistoryIDs []string
		rollback := func(cause error) {
			for i := len(createdHistoryIDs) - 1; i >= 0; i-- {
				_, _ = productStock.RemoveProductHistoryById(createdHistoryIDs[i])
			}
			for i := len(stockAdjustments) - 1; i >= 0; i-- {
				_, _ = productStock.AddProductStockQuantityById(stockAdjustments[i].stockId, stockAdjustments[i].quantity)
			}
			for i := len(soldFirstAdjustments) - 1; i >= 0; i-- {
				_, _ = productEntity.AddQuantitySoldFirstById(soldFirstAdjustments[i].productId, soldFirstAdjustments[i].quantity)
			}
			if orderCreated && result != nil {
				_, _ = orderEntity.RemoveOrderById(result.Id.Hex())
			}
			fields := logrus.Fields{
				"branchId": req.BranchId,
				"code":     req.Code,
			}
			if result != nil {
				fields["orderId"] = result.Id.Hex()
			}
			logrus.WithError(cause).WithFields(fields).Error("create order rolled back")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, cause.Error())
		}

		// Update product stock
		for _, item := range req.Items {
			if len(item.Stocks) > 0 {
				// Update stock quantity
				for _, itemStock := range item.Stocks {
					if itemStock.StockId != "" {
						stock, err := productStock.RemoveProductStockQuantityById(itemStock.StockId, itemStock.Quantity)
						if err != nil {
							rollback(fmt.Errorf("failed to update stock for product %s: %w", item.ProductId, err))
							return
						}
						if stock == nil {
							rollback(fmt.Errorf("failed to update stock for product %s", item.ProductId))
							return
						}
						updatedStocks = append(updatedStocks, *stock)
						stockAdjustments = append(stockAdjustments, stockAdjustment{
							stockId:  itemStock.StockId,
							quantity: itemStock.Quantity,
						})
					} else {
						if _, err := productEntity.RemoveQuantitySoldFirstById(item.ProductId, itemStock.Quantity); err != nil {
							rollback(fmt.Errorf("failed to update sold-first quantity for product %s: %w", item.ProductId, err))
							return
						}
						soldFirstAdjustments = append(soldFirstAdjustments, soldFirstAdjustment{
							productId: item.ProductId,
							quantity:  itemStock.Quantity,
						})
					}
				}

				// Add product history
				unit, err := productEntity.GetProductUnitById(item.UnitId)
				if err != nil || unit == nil {
					if err == nil {
						err = fmt.Errorf("product unit not found")
					}
					rollback(fmt.Errorf("failed to load unit for product %s: %w", item.ProductId, err))
					return
				}
				balance := productStock.GetProductStockBalance(item.ProductId, unit.Id.Hex(), req.BranchId)
				history := request.AddOrderItemProductHistory(item.ProductId, unit.Unit, item, balance, req.CreatedBy)
				history.BranchId = req.BranchId
				createdHistory, err := productStock.CreateProductHistory(history)
				if err != nil {
					rollback(fmt.Errorf("failed to create product history for product %s: %w", item.ProductId, err))
					return
				}
				if createdHistory != nil {
					createdHistoryIDs = append(createdHistoryIDs, createdHistory.Id.Hex())
				}
			}
		}

		ctx.JSON(http.StatusOK, gin.H{"data": result, "stocks": updatedStocks})
	}
}
