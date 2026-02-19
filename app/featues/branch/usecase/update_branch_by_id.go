package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func UpdateBranchById(branchEntity repositories.IBranch) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.Param("branchId")
		req := request.UpdateBranch{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId

		result, err := branchEntity.UpdateBranchById(branchId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
