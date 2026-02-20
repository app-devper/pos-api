package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"time"

	"github.com/gin-gonic/gin"
)

type getProductHistoryRangeQuery struct {
	StartDate time.Time `form:"startDate" binding:"required"`
	EndDate   time.Time `form:"endDate" binding:"required"`
}

func GetProductHistoryByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		branchId := ctx.GetString("BranchId")
		result, err := productEntity.GetProductHistoryByProductId(productId, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductHistoryByDateRange(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := getProductHistoryRangeQuery{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")
		result, err := productEntity.GetProductHistoryByDateRange(branchId, req.StartDate, req.EndDate)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
