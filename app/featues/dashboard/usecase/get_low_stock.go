package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetLowStockProducts(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		thresholdStr := ctx.DefaultQuery("threshold", "10")
		threshold, err := strconv.Atoi(thresholdStr)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_001, "invalid threshold")
			return
		}
		branchId := ctx.GetString("BranchId")
		result, err := productEntity.GetLowStockProducts(threshold, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
