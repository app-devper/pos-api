package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateStockAdjustment(
	stockAdjustmentEntity repositories.IStockAdjustment,
	productStock repositories.IProductStock,
	productEntity repositories.IProduct,
	orderEntity repositories.IOrder,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.StockAdjustment{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.AJ_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = utils.GetBranchId(ctx)

		adjustment, err := ApplyAdjustment(stockAdjustmentEntity, productStock, productEntity, orderEntity, sequenceEntity, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.AJ_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, adjustment)
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
