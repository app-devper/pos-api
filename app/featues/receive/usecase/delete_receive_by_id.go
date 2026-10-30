package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func DeleteReceiveById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("receiveId")
		if _, err := receiveEntity.GetReceiveById(id, utils.GetBranchId(ctx)); err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}
		result, err := receiveEntity.CancelReceiveById(id, utils.GetUserId(ctx))
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
