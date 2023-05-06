package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func GetProductById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("productId")
		result, err := productEntity.GetProductById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
