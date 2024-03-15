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
		product, err := productEntity.CreateProduct(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		unit, _ := productEntity.CreateProductUnitByProductId(product.Id.Hex(), req)
		_, _ = productEntity.CreateProductPriceByProductAndUnitId(product.Id.Hex(), unit.Id.Hex(), req)
		_, _ = productEntity.CreateProductStockByProductAndUnitId(product.Id.Hex(), unit.Id.Hex(), req)

		lot, _ := productEntity.CreateProductLotByProductId(product.Id.Hex(), req)
		if req.ReceiveId != "" {
			_, _ = receiveEntity.CreateReceiveItem(req.ReceiveId, lot.Id.Hex(), product.Id.Hex(), req)
		}

		ctx.JSON(http.StatusOK, product)
	}
}
