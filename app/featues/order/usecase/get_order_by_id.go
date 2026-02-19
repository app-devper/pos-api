package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetOrderById(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		result, err := orderEntity.GetOrderDetailById(orderId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
