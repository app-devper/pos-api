package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetMonthlyChart(orderEntity repositories.IOrderAnalytics) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := orderEntity.GetOrderMonthlyChart(branchId)
		if err != nil {
			logrus.WithError(err).WithField("branchId", branchId).Error("get dashboard monthly chart failed")
			errcode.AbortLookup(ctx, err, errcode.DA_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
