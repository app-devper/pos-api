package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	repositories "pos/app/data/repositories"
	"pos/app/domain/request"
)

func UpdateTotalCostById(orderEntity repositories.IOrder, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		order, err := orderEntity.GetOrderDetailById(orderId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		totalCost := 0.0
		for _, item := range order.Items {
			orderItem := request.OrderItem{
				CostPrice: productEntity.GetTotalCostPrice(item.ProductId.Hex(), item.Quantity),
				Quantity:  item.Quantity,
				Price:     item.Price,
				Discount:  item.Discount,
				ProductId: item.ProductId.Hex(),
			}
			_, _ = orderEntity.UpdateOrderItemById(item.Id.Hex(), orderItem)
			totalCost += orderItem.CostPrice
		}
		result, err := orderEntity.UpdateTotalCostOrderById(orderId, totalCost)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
