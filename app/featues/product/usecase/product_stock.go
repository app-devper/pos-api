package usecase

import (
	"errors"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreateProductStock(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductStock{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		req.BranchId = ctx.GetString("BranchId")
		// The Stock ledger writes its history and settles waiting Lines.
		stock, err := productStock.CreateProductStock(req)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.PD_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, stock)
	}
}

func GetProductStocksByProductId(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		branchId := ctx.GetString("BranchId")
		result, err := productStock.GetProductStocksByProductId(productId, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}

}

func UpdateProductStockById(productStock repositories.IProductStock, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStock{}
		id := ctx.Param("stockId")
		branchId := ctx.GetString("BranchId")
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		stock, err := productStock.GetProductStockById(id, branchId)
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.PD_BAD_REQUEST_002)
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		stock, err = productStock.UpdateProductStockById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		if err := appendProductStockHistory(branchId, productEntity, productStock, req.UnitId, func(unit *entities.ProductUnit, branchId string) request.ProductHistory {
			balance := productStock.GetProductStockBalance(req.ProductId, unit.Id.Hex(), branchId)
			updateHistory := request.UpdateProductStockHistory(req.ProductId, unit.Unit, req, balance)
			updateHistory.BranchId = branchId
			return updateHistory
		}); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, stock)
	}
}

func UpdateProductStockQuantityById(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStockQuantity{}
		id := ctx.Param("stockId")
		branchId := ctx.GetString("BranchId")
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		stock, err := productStock.GetProductStockById(id, branchId)
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.PD_BAD_REQUEST_002)
			return
		}
		// A one-Line Count in the Stock ledger: Adjustment, history, settlement.
		stock, err = productStock.UpdateProductStockQuantityById(id, branchId, req.Quantity, ctx.GetString("UserId"))
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.PD_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, stock)
	}
}

func RemoveProductStockById(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("stockId")
		userId := ctx.GetString("UserId")
		branchId := ctx.GetString("BranchId")
		if _, err := productStock.GetProductStockById(id, branchId); err != nil {
			errcode.AbortLookup(ctx, err, errcode.PD_BAD_REQUEST_002)
			return
		}

		result, err := productStock.RemoveProductStockById(id, branchId, userId)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.PD_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductStockSequence(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateProductStockSequence{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = ctx.GetString("BranchId")
		stocks, err := productStock.UpdateProductStockSequence(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, stocks)
	}
}

func appendProductStockHistory(branchId string, productEntity repositories.IProduct, productStock repositories.IProductStock, unitId string, buildHistory func(unit *entities.ProductUnit, branchId string) request.ProductHistory) error {
	unit, err := productEntity.GetProductUnitById(unitId)
	if err != nil {
		logrus.WithError(err).WithField("unitId", unitId).Error("failed to load product unit for stock history")
		return err
	}
	if unit == nil {
		logrus.WithField("unitId", unitId).Warn("product unit missing for stock history")
		return errors.New("product unit not found for stock history")
	}

	history := buildHistory(unit, branchId)
	if _, err = productStock.CreateProductHistory(history); err != nil {
		logrus.WithError(err).WithField("unitId", unitId).Error("failed to create product stock history")
		return err
	}
	return nil
}
