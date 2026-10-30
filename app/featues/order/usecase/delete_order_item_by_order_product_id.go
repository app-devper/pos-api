package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func DeleteOrderItemByOrderProductId(orderEntity repositories.IOrder, _ repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CancelOrderAction{}
		if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
			return
		}
		orderId := ctx.Param("orderId")
		productId := ctx.Param("productId")
		userId := ctx.GetString("UserId")
		if _, err := orderEntity.GetOrderById(orderId, ctx.GetString("BranchId")); err != nil {
			errcode.AbortLookup(ctx, err, errcode.OR_BAD_REQUEST_002)
			return
		}

		result, err := orderEntity.CancelOrderItemByOrderProductId(orderId, productId, userId, ctx.GetString("BranchId"), req.Reason)
		if err != nil {
			errcode.AbortLedger(ctx, err, errcode.OR_BAD_REQUEST_002)
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
