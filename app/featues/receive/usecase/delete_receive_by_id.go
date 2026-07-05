package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"

	"github.com/gin-gonic/gin"
)

func DeleteReceiveById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("receiveId")
		receive, err := receiveEntity.GetReceiveById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		if err := ensureReceiveBranchAccess(receive, utils.GetBranchId(ctx)); err != nil {
			abortReceiveBranchMismatch(ctx)
			return
		}
		if receive.Status == constant.IMPORTED {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, "cannot cancel imported receive")
			return
		}
		result, err := receiveEntity.UpdateReceiveStatusById(id, constant.CANCELLED, utils.GetUserId(ctx))
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
