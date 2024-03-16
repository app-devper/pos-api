package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	repository2 "pos/app/data/repository"
	"pos/app/domain/request"
)

func DeleteOrderItemByOrderProductId(orderEntity repository2.IOrder, productEntity repository2.IProduct) gin.HandlerFunc {
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

		if !result.StockId.IsZero() {
			_, _ = productEntity.AddProductStockQuantityById(result.StockId.Hex(), result.Quantity)

			// Add product history
			stock, _ := productEntity.GetProductStockById(result.StockId.Hex())
			unit, _ := productEntity.GetProductUnitById(stock.UnitId.Hex())
			balance := productEntity.GetProductStockBalance(result.ProductId.Hex(), unit.Id.Hex())
			_, _ = productEntity.CreateProductHistory(request.RemoveOrderItemProductHistory(result.ProductId.Hex(), unit.Unit, result, balance, userId))
		}

		date := utils.ToFormat(result.CreatedDate)
		_, _ = utils.NotifyMassage("ยกเลิกสินค้ารายการวันที่ " + date + "\n\n1. " + result.GetMessage())

		ctx.JSON(http.StatusOK, result)
	}
}
