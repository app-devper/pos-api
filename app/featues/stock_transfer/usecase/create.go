package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateStockTransfer(entity repositories.IStockTransfer, productEntity repositories.IProduct, sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.StockTransfer{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.FromBranchId = ctx.GetString("BranchId")

		sequence, _ := sequenceEntity.NextSequence(constant.STOCK_TRANSFER)
		if sequence != nil {
			req.Code = "TF-" + sequence.GenerateCode()
		}

		result, err := entity.CreateStockTransfer(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}

		for _, item := range req.Items {
			if item.StockId != "" {
				_, _ = productEntity.RemoveProductStockQuantityById(item.StockId, item.Quantity)
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}
