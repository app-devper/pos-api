package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetEmployeesByBranchId(employeeEntity repositories.IEmployee) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.Param("branchId")
		result, err := employeeEntity.GetEmployeesByBranchId(branchId)
		if err != nil {
			logrus.WithError(err).WithField("branchId", branchId).Error("get employees by branch id failed")
			errcode.AbortLookup(ctx, err, errcode.EM_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
