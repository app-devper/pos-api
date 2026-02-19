package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func DeleteBranchById(branchEntity repositories.IBranch) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.Param("branchId")
		result, err := branchEntity.RemoveBranchById(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
