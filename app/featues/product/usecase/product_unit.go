package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
)

func CreateProductUnit(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CreateProductUnit{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		productUnit := request.ProductUnit{
			ProductId:  req.ProductId,
			Unit:       req.Unit,
			Size:       req.Size,
			CostPrice:  req.CostPrice,
			Barcode:    req.Barcode,
			Volume:     req.Volume,
			VolumeUnit: req.VolumeUnit,
		}
		unit, _ := productEntity.CreateProductUnit(productUnit)

		productPrice := request.ProductPrice{
			ProductId:    req.ProductId,
			UnitId:       unit.Id.Hex(),
			Price:        req.Price,
			CustomerType: constant.CustomerTypeGeneral,
		}

		_, _ = productEntity.CreateProductPrice(productPrice)

		// Add product history
		_, _ = productEntity.CreateProductHistory(request.AddProductUnitHistory(req.ProductId, productUnit))

		ctx.JSON(http.StatusOK, unit)
	}
}

func GetProductUnitsByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.GetProductUnitsByProductId(productId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductUnitById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductUnit{}
		id := ctx.Param("id")
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		unit, err := productEntity.UpdateProductUnitById(id, req)

		// Add product history
		_, _ = productEntity.CreateProductHistory(request.UpdateProductUnitHistory(req.ProductId, req))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, unit)
	}
}

func RemoveProductUnitById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductUnitById(id)

		// Add product history
		_, _ = productEntity.CreateProductHistory(request.RemoveProductUnitHistory(result.ProductId.Hex(), result, userId))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_ = productEntity.RemoveProductPricesByUnitId(id)
		ctx.JSON(http.StatusOK, result)
	}
}
