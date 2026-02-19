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

func CreateBranch(
	branchEntity repositories.IBranch,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Branch{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		req.CreatedBy = userId

		sequence, _ := sequenceEntity.NextSequence(constant.BRANCH)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}

		result, err := branchEntity.CreateBranch(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
