package usecase

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
)

func DeleteOrderById(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		userId := ctx.GetString("UserId")
		result, err := orderEntity.RemoveOrderById(orderId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var message = ""
		var no = 1
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
			balance := productEntity.GetProductStockBalance(item.ProductId.Hex(), unit.Id.Hex())
			_, _ = productEntity.CreateProductHistory(request.RemoveOrderItemProductHistory(item.ProductId.Hex(), unit.Unit, &item, balance, userId))

			message += fmt.Sprintf("%d. %s\n", no, item.GetMessage())
			no += 1
		}
		message += fmt.Sprintf("\nรวม %.2f บาท", result.Total)

		date := utils.ToFormat(result.CreatedDate)
		_, _ = utils.NotifyMassage("ยกเลิกรายการวันที่ " + date + "\n\n" + message)

		ctx.JSON(http.StatusOK, result)
	}
}
