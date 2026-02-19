package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetOrdersByCustomerCode(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerCode := ctx.Param("customerCode")
		result, err := orderEntity.GetOrdersByCustomerCode(customerCode)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
