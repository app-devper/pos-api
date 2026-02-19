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

func CreateCreditNote(entity repositories.ICreditNote, productEntity repositories.IProduct, sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CreditNote{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CN_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = ctx.GetString("BranchId")
		sequence, _ := sequenceEntity.NextSequence(constant.CREDIT_NOTE)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}
		result, err := entity.CreateCreditNote(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CN_BAD_REQUEST_002, err.Error())
			return
		}

		for _, item := range req.Items {
			if item.StockId != "" {
				_, _ = productEntity.AddProductStockQuantityById(item.StockId, item.Quantity)
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}
