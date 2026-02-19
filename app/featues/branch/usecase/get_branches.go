package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetBranches(branchEntity repositories.IBranch) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := branchEntity.GetBranches()
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
