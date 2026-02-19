package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

func GetOrderItemDetailsByProductId(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetOrderRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			req.EndDate = time.Now()
		}
		productId := ctx.Param("productId")
		result, err := orderEntity.GetOrderItemOrderDetailsByProductId(productId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
