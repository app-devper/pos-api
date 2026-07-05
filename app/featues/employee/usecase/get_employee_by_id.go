package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetEmployeeById(employeeEntity repositories.IEmployee) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		employeeId := ctx.Param("employeeId")
		result, err := employeeEntity.GetEmployeeById(employeeId)
		if err != nil {
			logrus.WithError(err).WithField("employeeId", employeeId).Error("get employee by id failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.EM_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
