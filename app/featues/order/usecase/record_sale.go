package usecase

import (
	"context"
	"errors"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/ledger"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// recordSale records a Sale the till sent with its saleId. The Stock ledger
// prices it, draws its Stock and takes the Order code in one transaction; a
// repeat of a recorded Sale returns the Order already recorded.
// seller records a Sale through the Stock ledger.
type seller interface {
	Sell(ctx context.Context, form request.Sale) (*ledger.Sold, error)
}

func recordSale(ctx *gin.Context, sales seller) {
	req := request.Sale{}
	if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
		return
	}
	req.CreatedBy = utils.GetUserId(ctx)
	req.BranchId = utils.GetBranchId(ctx)

	recorded, err := sales.Sell(ctx.Request.Context(), req)
	var rejected *ledger.Rejected
	switch {
	case errors.Is(err, ledger.ErrSaleConflict):
		errcode.Abort(ctx, http.StatusConflict, errcode.OR_CONFLICT_001, "บิลนี้ถูกบันทึกไปแล้วด้วยรายการที่ต่างกัน ตรวจสอบประวัติการขายก่อนเริ่มบิลใหม่")
	case errors.As(err, &rejected):
		errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
	case err != nil:
		errcode.AbortLedger(ctx, err, errcode.OR_BAD_REQUEST_002)
	default:
		ctx.JSON(http.StatusOK, gin.H{"data": recorded.Order, "stocks": recorded.Stocks})
	}
}
