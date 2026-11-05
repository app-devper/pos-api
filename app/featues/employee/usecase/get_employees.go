package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetEmployees(employeeEntity repositories.IEmployee) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := employeeEntity.GetEmployees()
		if err != nil {
			logrus.WithError(err).Error("get employees failed")
			errcode.AbortLookup(ctx, err, errcode.EM_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
