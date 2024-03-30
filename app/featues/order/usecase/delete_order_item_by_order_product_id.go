package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
)

func DeleteOrderItemByOrderProductId(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		productId := ctx.Param("productId")
		userId := ctx.GetString("UserId")

		result, err := orderEntity.RemoveOrderItemByOrderProductId(orderId, productId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = orderEntity.UpdateTotalOrderById(orderId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, _ = productEntity.AddQuantityById(productId, result.Quantity)

		if len(result.Stocks) > 0 {
			for _, itemStock := range result.Stocks {
				_, _ = productEntity.AddProductStockQuantityById(itemStock.StockId, itemStock.Quantity)
			}
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		balance := productEntity.GetProductStockBalance(result.ProductId.Hex(), unit.Id.Hex())
		_, _ = productEntity.CreateProductHistory(request.RemoveOrderItemProductHistory(result.ProductId.Hex(), unit.Unit, result, balance, userId))

		date := utils.ToFormat(result.CreatedDate)
		_, _ = utils.NotifyMassage("ยกเลิกสินค้ารายการวันที่ " + date + "\n\n1. " + result.GetMessage())

		ctx.JSON(http.StatusOK, result)
	}
}
