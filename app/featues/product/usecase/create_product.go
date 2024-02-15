package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
	"pos/app/domain/request"
)

func CreateProduct(productEntity repository.IProduct, receiveEntity repository.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Product{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.CreatedBy = userId
		result, err := productEntity.CreateProduct(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		lot, _ := productEntity.CreateProductLotByProductId(result.Id.Hex(), req)
		if req.ReceiveId != "" {
			_, _ = receiveEntity.CreateReceiveItem(req.ReceiveId, lot.Id.Hex(), result.Id.Hex(), req)
		}

		ctx.JSON(http.StatusOK, result)
	}
}
