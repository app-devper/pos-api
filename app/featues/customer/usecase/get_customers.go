package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetCustomers(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := customerEntity.GetCustomerAll()
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
