package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetABCAnalysis(orderEntity repositories.IOrderAnalytics) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := orderEntity.GetABCAnalysis(branchId)
		if err != nil {
			logrus.WithError(err).WithField("branchId", branchId).Error("get abc analysis failed")
			errcode.AbortLookup(ctx, err, errcode.DA_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
