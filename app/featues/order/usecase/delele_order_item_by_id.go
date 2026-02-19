package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func DeleteOrderItemById(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		itemId := ctx.Param("itemId")
		userId := ctx.GetString("UserId")
		result, err := orderEntity.RemoveOrderItemById(itemId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}
		_, err = orderEntity.UpdateTotalOrderById(result.OrderId.Hex())
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		if len(result.Stocks) > 0 {
			// Update stock quantity
			for _, itemStock := range result.Stocks {
				if itemStock.StockId != "" {
					_, _ = productEntity.AddProductStockQuantityById(itemStock.StockId, itemStock.Quantity)
				} else {
					_, _ = productEntity.AddQuantitySoldFirstById(result.ProductId.Hex(), itemStock.Quantity)
				}
			}

			// Add product history
			unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
			if unit != nil {
				balance := productEntity.GetProductStockBalance(result.ProductId.Hex(), unit.Id.Hex())
				_, _ = productEntity.CreateProductHistory(request.RemoveOrderItemProductHistory(result.ProductId.Hex(), unit.Unit, result, balance, userId))
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}
