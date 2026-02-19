package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func DeleteOrderById(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		userId := ctx.GetString("UserId")
		result, err := orderEntity.RemoveOrderById(orderId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		for _, item := range result.Items {

			if len(item.Stocks) > 0 {
				// Update stock quantity
				for _, itemStock := range item.Stocks {
					if itemStock.StockId != "" {
						_, _ = productEntity.AddProductStockQuantityById(itemStock.StockId, itemStock.Quantity)
					} else {
						_, _ = productEntity.AddQuantitySoldFirstById(item.ProductId.Hex(), itemStock.Quantity)
					}
				}
			}

			// Add product history
			unit, _ := productEntity.GetProductUnitById(item.UnitId.Hex())
			if unit != nil {
				balance := productEntity.GetProductStockBalance(item.ProductId.Hex(), unit.Id.Hex())
				_, _ = productEntity.CreateProductHistory(request.RemoveOrderItemProductHistory(item.ProductId.Hex(), unit.Unit, &item, balance, userId))
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}
