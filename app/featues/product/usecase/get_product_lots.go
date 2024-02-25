package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
	"pos/app/domain/request"
)

func GetProductLots(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetExpireRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := productEntity.GetProductLots(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
