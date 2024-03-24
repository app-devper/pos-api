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
		unit, err := productEntity.CreateProductUnit(productUnit)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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
		id := ctx.Param("unitId")
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		unit, err := productEntity.UpdateProductUnitById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Add product history
		_, _ = productEntity.CreateProductHistory(request.UpdateProductUnitHistory(req.ProductId, req))

		ctx.JSON(http.StatusOK, unit)
	}
}

func RemoveProductUnitById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("unitId")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductUnitById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Add product history
		_, _ = productEntity.CreateProductHistory(request.RemoveProductUnitHistory(result.ProductId.Hex(), result, userId))
		_ = productEntity.RemoveProductPricesByUnitId(id)
		ctx.JSON(http.StatusOK, result)
	}
}
