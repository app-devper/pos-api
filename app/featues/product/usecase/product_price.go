package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
)

func CreateProductPrice(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductPrice{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customerTypes := constant.CustomerTypes()
		if customerTypeIsValid := utils.InArrayString(req.CustomerType, customerTypes); !customerTypeIsValid {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "customer type is not valid"})
			return
		}

		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		result, err := productEntity.CreateProductPrice(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		_, _ = productEntity.CreateProductHistory(request.AddProductPriceHistory(req.ProductId, unit.Unit, req))

		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductPricesByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
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

func UpdateProductPriceById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductPrice{}
		id := ctx.Param("priceId")
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		customerTypes := constant.CustomerTypes()
		if customerTypeIsValid := utils.InArrayString(req.CustomerType, customerTypes); !customerTypeIsValid {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "customer type is not valid"})
			return
		}

		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		result, err := productEntity.UpdateProductPriceById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(req.UnitId)
		_, _ = productEntity.CreateProductHistory(request.UpdateProductPriceHistory(req.ProductId, unit.Unit, req))

		ctx.JSON(http.StatusOK, result)
	}
}

func RemoveProductPriceById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("priceId")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductPriceById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		_, _ = productEntity.CreateProductHistory(request.RemoveProductPriceHistory(result.ProductId.Hex(), unit.Unit, result, userId))

		ctx.JSON(http.StatusOK, result)
	}
}
