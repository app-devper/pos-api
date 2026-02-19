package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func UpdateCustomerCodeOrderById(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateCustomerCode{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
			return
		}
		orderId := ctx.Param("orderId")
		result, err := orderEntity.UpdateCustomerCodeOrderById(orderId, req.CustomerCode)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
