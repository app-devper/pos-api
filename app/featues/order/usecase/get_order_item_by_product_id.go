package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
)

func GetOrderItemByProductId(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := orderEntity.GetOrderItemByProductId(productId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
