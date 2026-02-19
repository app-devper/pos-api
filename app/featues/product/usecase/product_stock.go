package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateProductStock(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductStock{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		req.BranchId = ctx.GetString("BranchId")
		stock, err := productEntity.CreateProductStock(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		// Add product history
		unit, _ := productEntity.GetProductUnitById(req.UnitId)
		if unit != nil {
			balance := productEntity.GetProductStockBalance(req.ProductId, unit.Id.Hex())
			history := request.AddProductStockHistory(req.ProductId, unit.Unit, req, balance)
			history.BranchId = req.BranchId
			_, _ = productEntity.CreateProductHistory(history)
		}

		ctx.JSON(http.StatusOK, stock)
	}
}

func GetProductStocksByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.GetProductStocksByProductId(productId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}

}

func UpdateProductStockById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStock{}
		id := ctx.Param("stockId")
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		stock, err := productEntity.UpdateProductStockById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(req.UnitId)
		if unit != nil {
			balance := productEntity.GetProductStockBalance(req.ProductId, unit.Id.Hex())
			updateHistory := request.UpdateProductStockHistory(req.ProductId, unit.Unit, req, balance)
			updateHistory.BranchId = ctx.GetString("BranchId")
			_, _ = productEntity.CreateProductHistory(updateHistory)
		}

		ctx.JSON(http.StatusOK, stock)
	}
}

func UpdateProductStockQuantityById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStockQuantity{}
		id := ctx.Param("stockId")
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		stock, err := productEntity.UpdateProductStockQuantityById(id, req.Quantity)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		// Add product history
		unit, _ := productEntity.GetProductUnitById(stock.UnitId.Hex())
		if unit != nil {
			balance := productEntity.GetProductStockBalance(stock.ProductId.Hex(), unit.Id.Hex())
			qtyHistory := request.UpdateProductStockQuantityHistory(stock.ProductId.Hex(), unit.Unit, req, balance)
			qtyHistory.BranchId = ctx.GetString("BranchId")
			_, _ = productEntity.CreateProductHistory(qtyHistory)
		}

		ctx.JSON(http.StatusOK, stock)
	}
}

func RemoveProductStockById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("stockId")
		userId := ctx.GetString("UserId")

		result, err := productEntity.RemoveProductStockById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		// Add product history
		unit, _ := productEntity.GetProductUnitById(result.UnitId.Hex())
		if unit != nil {
			balance := productEntity.GetProductStockBalance(result.ProductId.Hex(), unit.Id.Hex())
			removeHistory := request.RemoveProductStockHistory(result.ProductId.Hex(), unit.Unit, result, balance, userId)
			removeHistory.BranchId = ctx.GetString("BranchId")
			_, _ = productEntity.CreateProductHistory(removeHistory)
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductStockSequence(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStockSequence{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		stocks, err := productEntity.UpdateProductStockSequence(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, stocks)
	}
}
