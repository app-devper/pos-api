package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetOrderItemByProductId(orderEntity repositories.IOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := orderEntity.GetOrderItemByProductId(productId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
