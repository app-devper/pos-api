package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func DeleteOrderItemByOrderProductId(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		productId := ctx.Param("productId")
		userId := ctx.GetString("UserId")

		result, err := orderEntity.RemoveOrderItemByOrderProductId(orderId, productId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		_, err = orderEntity.UpdateTotalOrderById(orderId)
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
					_, _ = productEntity.AddQuantitySoldFirstById(productId, itemStock.Quantity)
				}
			}
		}

		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		if unit != nil {
			balance := productEntity.GetProductStockBalance(result.ProductId.Hex(), unit.Id.Hex())
			h := request.RemoveOrderItemProductHistory(result.ProductId.Hex(), unit.Unit, result, balance, userId)
			h.BranchId = ctx.GetString("BranchId")
			_, _ = productEntity.CreateProductHistory(h)
		}

		ctx.JSON(http.StatusOK, result)
	}
}
