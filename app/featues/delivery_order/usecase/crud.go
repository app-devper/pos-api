package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetDeliveryOrders(entity repositories.IDeliveryOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetDeliveryOrders(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetDeliveryOrderById(entity repositories.IDeliveryOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetDeliveryOrderById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateDeliveryOrderById(entity repositories.IDeliveryOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateDeliveryOrder{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DO_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdateDeliveryOrderById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteDeliveryOrderById(entity repositories.IDeliveryOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.RemoveDeliveryOrderById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
