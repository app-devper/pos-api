package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func DeleteOrderItemByOrderProductId(orderEntity repositories.IOrder, _ repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		productId := ctx.Param("productId")
		userId := ctx.GetString("UserId")
		order, err := orderEntity.GetOrderById(orderId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}
		if err := ensureOrderBranchAccess(order, ctx.GetString("BranchId")); err != nil {
			abortOrderBranchMismatch(ctx)
			return
		}

		result, err := orderEntity.CancelOrderItemByOrderProductId(orderId, productId, userId, ctx.GetString("BranchId"))
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
