package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func ImportReceiveToStock(receiveEntity repositories.IReceive, _ repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		receiveId := ctx.Param("receiveId")

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)
		receive, err := receiveEntity.GetReceiveById(receiveId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		if err := ensureReceiveBranchAccess(receive, branchId); err != nil {
			abortReceiveBranchMismatch(ctx)
			return
		}
		result, err := receiveEntity.ImportReceiveToStock(receiveId, userId, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
