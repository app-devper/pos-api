package usecase

import (
	"errors"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// recordSale records a Sale the till sent with its saleId: the server prices
// it and draws its Stock. A repeat of a recorded Sale returns the
// Order already recorded, without consuming an Order code.
func recordSale(ctx *gin.Context, orderEntity repositories.IOrder, sequenceEntity repositories.ISequence) {
	req := request.Sale{}
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
		return
	}
	req.CreatedBy = utils.GetUserId(ctx)
	req.BranchId = utils.GetBranchId(ctx)

	respond := func(recorded *repositories.RecordedSale, err error) bool {
		var rejected *repositories.SaleRejected
		switch {
		case errors.Is(err, repositories.ErrSaleConflict):
			errcode.Abort(ctx, http.StatusConflict, errcode.OR_CONFLICT_001, err.Error())
		case errors.As(err, &rejected):
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
		case err != nil:
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
		case recorded != nil:
			ctx.JSON(http.StatusOK, gin.H{"data": recorded.Order, "stocks": recorded.Stocks})
		default:
			return false
		}
		return true
	}

	if respond(orderEntity.FindSale(req)) {
		return
	}
	sequence, err := sequenceEntity.NextSequence(constant.ORDER)
	if err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
		return
	}
	if sequence == nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, "order sequence not available")
		return
	}
	req.Code = sequence.GenerateCode()
	respond(orderEntity.RecordSale(req))
}
