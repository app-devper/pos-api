package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
)

func DeleteOrderItemById(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		itemId := ctx.Param("itemId")
		userId := ctx.GetString("UserId")
		result, err := orderEntity.RemoveOrderItemById(itemId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err = orderEntity.UpdateTotalOrderById(result.OrderId.Hex())
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, _ = productEntity.AddQuantityById(result.ProductId.Hex(), result.Quantity)

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
