package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetStockReport(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := productStock.GetStockReport(branchId)
		if err != nil {
			logrus.WithError(err).WithField("branchId", branchId).Error("get dashboard stock report failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
