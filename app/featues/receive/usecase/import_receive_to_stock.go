package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

// ImportReceiveToStock moves a Receive into Stock. Oversell settlement happens
// inside the same transaction, in the Stock ledger (ADR-0002).
func ImportReceiveToStock(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		receiveId := ctx.Param("receiveId")

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)
		if _, err := receiveEntity.GetReceiveById(receiveId, branchId); err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
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
