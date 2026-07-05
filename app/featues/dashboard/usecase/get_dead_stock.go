package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetDeadStockProducts(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		daysStr := ctx.DefaultQuery("days", "90")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			logrus.WithError(err).WithField("days", daysStr).Error("parse dead stock days failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_001, "invalid days")
			return
		}
		branchId := ctx.GetString("BranchId")
		result, err := productStock.GetDeadStockProducts(days, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId": branchId,
				"days":     days,
			}).Error("get dead stock products failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
