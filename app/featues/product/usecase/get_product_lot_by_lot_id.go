package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func GetProductLotByLotId(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lotId := ctx.Param("lotId")

		result, err := productEntity.GetProductLotById(lotId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
