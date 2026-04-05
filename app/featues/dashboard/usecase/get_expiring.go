package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetExpiringProducts(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		now := time.Now()
		sixMonthsLater := now.AddDate(0, 6, 0)
		param := request.GetProductLotsExpireRange{
			StartDate: now,
			EndDate:   sixMonthsLater,
		}
		branchId := ctx.GetString("BranchId")
		result, err := productEntity.GetExpiringProductStocks(param, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":  branchId,
				"startDate": param.StartDate,
				"endDate":   param.EndDate,
			}).Error("get expiring products failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
