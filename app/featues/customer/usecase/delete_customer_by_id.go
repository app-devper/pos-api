package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func DeleteCustomerById(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerId := ctx.Param("customerId")
		result, err := customerEntity.RemoveCustomerById(customerId)
		if err != nil {
			logrus.WithError(err).WithField("customerId", customerId).Error("delete customer failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
