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

// adjuster records an Adjustment through the Stock ledger.
type adjuster interface {
	Adjust(ctx context.Context, req request.StockAdjustment) (*entities.StockAdjustment, error)
}

func CreateStockAdjustment(ledger adjuster) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.StockAdjustment{}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.AJ_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = utils.GetBranchId(ctx)
		req.CreatedBy = utils.GetUserId(ctx)
		result, err := ledger.Adjust(ctx.Request.Context(), req)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.AJ_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetStockAdjustmentsByProductId(stockAdjustmentEntity repositories.IStockAdjustment) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		branchId := utils.GetBranchId(ctx)
		result, err := stockAdjustmentEntity.GetStockAdjustmentsByProductId(productId, branchId)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.AJ_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
