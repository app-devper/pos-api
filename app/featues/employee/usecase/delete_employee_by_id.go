package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func DeleteEmployeeById(employeeEntity repositories.IEmployee) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		employeeId := ctx.Param("employeeId")
		result, err := employeeEntity.RemoveEmployeeById(employeeId, utils.GetUserId(ctx))
		if err != nil {
			logrus.WithError(err).WithField("employeeId", employeeId).Error("delete employee failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.EM_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
