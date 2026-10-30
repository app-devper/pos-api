package usecase

import (
	"context"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

// countRecorder records a Count through the Stock ledger.
type countRecorder interface {
	Count(ctx context.Context, req request.StockCount) (*entities.StockCount, error)
}

func CreateStockCount(ledger countRecorder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.StockCount{}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = utils.GetBranchId(ctx)
		req.CreatedBy = utils.GetUserId(ctx)
		result, err := ledger.Count(ctx.Request.Context(), req)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.SC_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetStockCounts(stockCountEntity repositories.IStockCount) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := utils.GetBranchId(ctx)
		result, err := stockCountEntity.GetStockCounts(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SC_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetStockCountById(stockCountEntity repositories.IStockCount) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := stockCountEntity.GetStockCountById(id, utils.GetBranchId(ctx))
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.SC_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
