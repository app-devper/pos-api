package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetReceivesRange(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetReceiveRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = ctx.GetString("BranchId")
		result, err := receiveEntity.GetReceives(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
