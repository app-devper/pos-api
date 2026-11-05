package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetCustomerById(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerId := ctx.Param("customerId")
		result, err := customerEntity.GetCustomerById(customerId)
		if err != nil {
			logrus.WithError(err).WithField("customerId", customerId).Error("get customer by id failed")
			errcode.AbortLookup(ctx, err, errcode.CU_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
