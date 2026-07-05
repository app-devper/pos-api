package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetReceiveById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("receiveId")
		result, err := receiveEntity.GetReceiveById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		if err := ensureReceiveBranchAccess(result, utils.GetBranchId(ctx)); err != nil {
			abortReceiveBranchMismatch(ctx)
			return
		}
		items, err := receiveEntity.GetReceiveItemsByReceiveId(id)
		if err == nil && len(items) > 0 {
			result.Items = items
		}
		ctx.JSON(http.StatusOK, result)
	}
}
