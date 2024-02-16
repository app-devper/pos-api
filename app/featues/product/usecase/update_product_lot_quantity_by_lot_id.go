package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
	"pos/app/domain/request"
)

func UpdateProductLotQuantityByLotId(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("lotId")
		req := request.UpdateProductLotQuantity{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		result, err := productEntity.UpdateProductLotQuantityById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
