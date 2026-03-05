package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetDeadStockProducts(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		daysStr := ctx.DefaultQuery("days", "90")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_001, "invalid days")
			return
		}
		branchId := ctx.GetString("BranchId")
		result, err := productEntity.GetDeadStockProducts(days, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
