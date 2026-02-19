package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func UpdateReceiveById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceive{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		id := ctx.Param("receiveId")

		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId

		result, err := receiveEntity.UpdateReceiveById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateReceiveItemsById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceiveItems{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		receiveId := ctx.Param("receiveId")

		result, err := receiveEntity.UpdateReceiveItemsById(receiveId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateReceiveTotalCostById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceiveTotalCode{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		id := ctx.Param("receiveId")

		result, err := receiveEntity.UpdateReceiveTotalCostById(id, req.TotalCost)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
