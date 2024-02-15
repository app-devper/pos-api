package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func GetProductLotsExpired(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := productEntity.GetProductLotsExpired()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
