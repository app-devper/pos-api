package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetPurchaseOrders(entity repositories.IPurchaseOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPurchaseOrders(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetPurchaseOrderById(entity repositories.IPurchaseOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetPurchaseOrderById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdatePurchaseOrderById(entity repositories.IPurchaseOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdatePurchaseOrder{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PO_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdatePurchaseOrderById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeletePurchaseOrderById(entity repositories.IPurchaseOrder) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.RemovePurchaseOrderById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PO_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
