package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func DeleteBranchById(branchEntity repositories.IBranch) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.Param("branchId")
		result, err := branchEntity.RemoveBranchById(branchId, utils.GetUserId(ctx))
		if err != nil {
			logrus.WithError(err).WithField("branchId", branchId).Error("delete branch failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
