package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func DeleteCustomerById(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerId := ctx.Param("customerId")
		result, err := customerEntity.RemoveCustomerById(customerId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
