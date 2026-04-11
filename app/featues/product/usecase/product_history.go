package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

type getProductHistoryRangeQuery struct {
	StartDate request.FlexibleTime `form:"startDate" binding:"required"`
	EndDate   request.FlexibleTime `form:"endDate" binding:"required"`
}

func GetProductHistoryByProductId(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		branchId := ctx.GetString("BranchId")
		result, err := productStock.GetProductHistoryByProductId(productId, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductHistoryByDateRange(productStock repositories.IProductStock) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := getProductHistoryRangeQuery{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		result, err := productStock.GetProductHistoryByDateRange(branchId, req.StartDate.Time, req.EndDate.Time)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
