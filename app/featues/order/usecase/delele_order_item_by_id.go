package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func DeleteOrderItemById(orderEntity repositories.IOrder, _ repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CancelOrderAction{}
		_ = ctx.ShouldBindJSON(&req)
		itemId := ctx.Param("itemId")
		userId := ctx.GetString("UserId")
		item, err := orderEntity.GetOrderItemById(itemId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}
		if err := ensureOrderItemBranchAccess(item, ctx.GetString("BranchId")); err != nil {
			abortOrderBranchMismatch(ctx)
			return
		}
		result, err := orderEntity.CancelOrderItemById(itemId, userId, ctx.GetString("BranchId"), req.Reason)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
