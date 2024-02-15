package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func DeleteReceiveItemByLotId(receiveEntity repository.IReceive, productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("lotId")
		result, err := receiveEntity.RemoveReceiveItemByLotId(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, _ = productEntity.RemoveProductLotById(id)
		_, _ = productEntity.RemoveQuantityById(result.ProductId.Hex(), result.Quantity)

		ctx.JSON(http.StatusOK, result)
	}
}
