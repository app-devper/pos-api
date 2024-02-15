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
		lot, err := productEntity.RemoveProductLotById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err = productEntity.RemoveQuantityById(lot.ProductId.Hex(), lot.Quantity)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
