package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
	"pos/app/domain/request"
)

func CreateProductPrice(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductPrice{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		result, err := productEntity.CreateProductPrice(req)

		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		_, _ = productEntity.CreateProductHistory(request.AddProductPriceHistory(req.ProductId, unit.Unit, req))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductPricesByProductId(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.GetProductPricesByProductId(productId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductPriceById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductPrice{}
		id := ctx.Param("id")
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		result, err := productEntity.UpdateProductPriceById(id, req)
		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		_, _ = productEntity.CreateProductHistory(request.UpdateProductPriceHistory(result.ProductId.Hex(), unit.Unit, req))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func RemoveProductPriceById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductPriceById(id)

		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		_, _ = productEntity.CreateProductHistory(request.RemoveProductPriceHistory(result.ProductId.Hex(), unit.Unit, result, userId))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
