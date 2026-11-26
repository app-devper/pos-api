package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

// excludeCancelledParam opts the list out of cancelled Receives; without it
// the list is every Receive in the range, cancelled ones included.
const excludeCancelledParam = "excludeCancelled"

func GetReceivesRange(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetReceiveRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		excludeCancelled, ok := request.BoolQuery(ctx, excludeCancelledParam)
		if !ok {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, excludeCancelledParam+" must be true or false")
			return
		}
		req.ExcludeCancelled = excludeCancelled
		req.BranchId = ctx.GetString("BranchId")
		result, err := receiveEntity.GetReceives(req)
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
