package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateOrder(
	orderEntity repositories.IOrder,
	productEntity repositories.IProduct,
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

		sequence, _ := sequenceEntity.NextSequence(constant.ORDER)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}

		result, err := orderEntity.CreateOrder(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		// Update product stock
		var stocks []entities.ProductStock
		for _, item := range req.Items {
			if len(item.Stocks) > 0 {
				// Update stock quantity
				for _, itemStock := range item.Stocks {
					if itemStock.StockId != "" {
						stock, err := productEntity.RemoveProductStockQuantityById(itemStock.StockId, itemStock.Quantity)
						if err == nil && stock != nil {
							stocks = append(stocks, *stock)
						}
					} else {
						_, _ = productEntity.RemoveQuantitySoldFirstById(item.ProductId, itemStock.Quantity)
					}
				}

				// Add product history
				unit, _ := productEntity.GetProductUnitById(item.UnitId)
				if unit != nil {
					balance := productEntity.GetProductStockBalance(item.ProductId, unit.Id.Hex())
					history := request.AddOrderItemProductHistory(item.ProductId, unit.Unit, item, balance, req.CreatedBy)
					history.BranchId = req.BranchId
					_, _ = productEntity.CreateProductHistory(history)
				}
			}
		}

		ctx.JSON(http.StatusOK, gin.H{"data": result, "stocks": stocks})
	}
}
