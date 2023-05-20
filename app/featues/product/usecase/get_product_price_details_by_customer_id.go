package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func GetProductPriceDetailsByCustomerId(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerId := ctx.Param("customerId")
		result, err := productEntity.GetProductPriceDetailByCustomerId(customerId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
