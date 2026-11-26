package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

// excludeCancelledParam opts the list out of cancelled Orders; without it the
// list is every Order in the range, cancelled ones included.
const excludeCancelledParam = "excludeCancelled"

func GetOrdersRange(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetOrderRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
			return
		}
		excludeCancelled, ok := request.BoolQuery(ctx, excludeCancelledParam)
		if !ok {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, excludeCancelledParam+" must be true or false")
			return
		}
		req.ExcludeCancelled = excludeCancelled
		req.BranchId = ctx.GetString("BranchId")
		result, err := orderEntity.GetOrderRange(req)
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.OR_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
