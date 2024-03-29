package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
)

func GetOrderById(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		result, err := orderEntity.GetOrderDetailById(orderId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
