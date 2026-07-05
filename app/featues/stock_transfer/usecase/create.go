package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
		if req.ToBranchId == req.FromBranchId {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_001, "destination branch must be different from source branch")
			return
		}

		sequence, err := sequenceEntity.NextSequence(constant.STOCK_TRANSFER)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}
		if sequence == nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, "stock transfer sequence not available")
			return
		}
		req.Code = "TF-" + sequence.GenerateCode()

		result, err := entity.CreateStockTransferWithReservation(req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"fromBranchId": req.FromBranchId,
				"toBranchId":   req.ToBranchId,
				"code":         req.Code,
			}).Error("create stock transfer failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.TR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
