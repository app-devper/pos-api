package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateCustomerHistory(entity repositories.ICustomerHistory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CustomerHistory{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CH_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = ctx.GetString("BranchId")
		result, err := entity.CreateCustomerHistory(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CH_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetCustomerHistories(entity repositories.ICustomerHistory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerCode := ctx.Param("customerCode")
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetCustomerHistories(customerCode, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CH_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
