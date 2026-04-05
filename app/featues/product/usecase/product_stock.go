package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

		appendProductStockHistory(req.BranchId, productEntity, req.UnitId, func(unit *entities.ProductUnit, branchId string) request.ProductHistory {
			balance := productEntity.GetProductStockBalance(req.ProductId, unit.Id.Hex(), branchId)
			history := request.AddProductStockHistory(req.ProductId, unit.Unit, req, balance)
			history.BranchId = branchId
			return history
		})

		ctx.JSON(http.StatusOK, stock)
	}
}

func GetProductStocksByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		branchId := ctx.GetString("BranchId")
		result, err := productEntity.GetProductStocksByProductId(productId, branchId)
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
		appendProductStockHistory(ctx.GetString("BranchId"), productEntity, req.UnitId, func(unit *entities.ProductUnit, branchId string) request.ProductHistory {
			balance := productEntity.GetProductStockBalance(req.ProductId, unit.Id.Hex(), branchId)
			updateHistory := request.UpdateProductStockHistory(req.ProductId, unit.Unit, req, balance)
			updateHistory.BranchId = branchId
			return updateHistory
		})

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
		appendProductStockHistory(ctx.GetString("BranchId"), productEntity, stock.UnitId.Hex(), func(unit *entities.ProductUnit, branchId string) request.ProductHistory {
			balance := productEntity.GetProductStockBalance(stock.ProductId.Hex(), unit.Id.Hex(), branchId)
			qtyHistory := request.UpdateProductStockQuantityHistory(stock.ProductId.Hex(), unit.Unit, req, balance)
			qtyHistory.BranchId = branchId
			return qtyHistory
		})

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

		appendProductStockHistory(ctx.GetString("BranchId"), productEntity, result.UnitId.Hex(), func(unit *entities.ProductUnit, branchId string) request.ProductHistory {
			balance := productEntity.GetProductStockBalance(result.ProductId.Hex(), unit.Id.Hex(), branchId)
			removeHistory := request.RemoveProductStockHistory(result.ProductId.Hex(), unit.Unit, result, balance, userId)
			removeHistory.BranchId = branchId
			return removeHistory
		})

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
		req.BranchId = ctx.GetString("BranchId")
		stocks, err := productEntity.UpdateProductStockSequence(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, stocks)
	}
}

func appendProductStockHistory(branchId string, productEntity repositories.IProduct, unitId string, buildHistory func(unit *entities.ProductUnit, branchId string) request.ProductHistory) {
	unit, err := productEntity.GetProductUnitById(unitId)
	if err != nil {
		logrus.WithError(err).WithField("unitId", unitId).Error("failed to load product unit for stock history")
		return
	}
	if unit == nil {
		logrus.WithField("unitId", unitId).Warn("product unit missing for stock history")
		return
	}

	history := buildHistory(unit, branchId)
	if _, err = productEntity.CreateProductHistory(history); err != nil {
		logrus.WithError(err).WithField("unitId", unitId).Error("failed to create product stock history")
	}
}
