package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateReceive(receiveEntity repositories.IReceive, sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Receive{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)

		sequence, _ := sequenceEntity.NextSequence(constant.RECEIVE)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}
		req.UpdatedBy = userId
		req.BranchId = utils.GetBranchId(ctx)

		result, err := receiveEntity.CreateReceive(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
