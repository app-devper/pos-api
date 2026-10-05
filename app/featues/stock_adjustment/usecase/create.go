package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
)

func CreateStockAdjustment(records repositories.IStockAdjustment) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.StockAdjustment{}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.AJ_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = utils.GetBranchId(ctx)
		req.CreatedBy = utils.GetUserId(ctx)
		result, err := records.ApplyStockAdjustment(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.AJ_BAD_REQUEST_002, err.Error())
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
			errcode.Abort(ctx, http.StatusBadRequest, errcode.AJ_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
