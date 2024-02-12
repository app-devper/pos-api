package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func GetReceiveItemByReceiveId(receiveEntity repository.IReceive, productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("receiveId")
		items, err := receiveEntity.GetReceiveItemsByReceiveId(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ids := make([]string, 0, len(items))
		for _, value := range items {
			ids = append(ids, value.LotId.Hex())
		}
		result, err := productEntity.GetProductLotsByIds(ids)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
