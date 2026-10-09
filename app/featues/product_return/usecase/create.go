package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateProductReturn(records repositories.IProductReturn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductReturn{}
		if err := ctx.ShouldBindJSON(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = utils.GetBranchId(ctx)
		req.CreatedBy = utils.GetUserId(ctx)
		result, err := records.RecordProductReturn(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductReturnsByOrderId(productReturnEntity repositories.IProductReturn) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		orderId := ctx.Param("orderId")
		branchId := utils.GetBranchId(ctx)
		result, err := productReturnEntity.GetProductReturnsByOrderId(orderId, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
