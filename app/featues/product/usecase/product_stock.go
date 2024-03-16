package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repository"
	"pos/app/domain/request"
)

func CreateProductStock(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductStock{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		stock, err := productEntity.CreateProductStock(req)

		// Add product history
		unit, _ := productEntity.GetProductUnitById(req.UnitId)
		balance := productEntity.GetProductStockBalance(req.ProductId, unit.Id.Hex())
		_, _ = productEntity.CreateProductHistory(request.AddProductStockHistory(req.ProductId, unit.Unit, req, balance))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, stock)
	}
}

func GetProductStocksByProductId(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.GetProductStocksByProductId(productId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}

}

func UpdateProductStockById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStock{}
		id := ctx.Param("id")
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		stock, err := productEntity.UpdateProductStockById(id, req)

		// Add product history
		unit, _ := productEntity.GetProductUnitById(req.UnitId)
		balance := productEntity.GetProductStockBalance(req.ProductId, unit.Id.Hex())
		_, _ = productEntity.CreateProductHistory(request.UpdateProductStockHistory(req.ProductId, unit.Unit, req, balance))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, stock)
	}
}

func UpdateProductStockQuantityById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStockQuantity{}
		id := ctx.Param("id")
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		stock, err := productEntity.UpdateProductStockQuantityById(id, req.Quantity)

		// Add product history
		unit, _ := productEntity.GetProductUnitById(stock.UnitId.Hex())
		balance := productEntity.GetProductStockBalance(stock.ProductId.Hex(), unit.Id.Hex())
		_, _ = productEntity.CreateProductHistory(request.UpdateProductStockQuantityHistory(stock.ProductId.Hex(), unit.Unit, req, balance))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, stock)
	}
}

func RemoveProductStockById(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductStockById(id)

		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		balance := productEntity.GetProductStockBalance(result.ProductId.Hex(), unit.Id.Hex())
		_, _ = productEntity.CreateProductHistory(request.RemoveProductStockHistory(result.ProductId.Hex(), unit.Unit, result, balance, userId))

		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductStockSequence(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStockSequence{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		stocks, err := productEntity.UpdateProductStockSequence(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, stocks)
	}
}
