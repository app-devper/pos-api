package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repository"
)

func GetOrderItemById(orderEntity repository.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		itemId := ctx.Param("itemId")
		result, err := orderEntity.GetOrderItemDetailById(itemId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
