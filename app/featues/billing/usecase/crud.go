package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetBillings(entity repositories.IBilling) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetBillings(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BL_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetBillingById(entity repositories.IBilling) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetBillingById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BL_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateBillingById(entity repositories.IBilling) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateBilling{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BL_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdateBillingById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BL_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteBillingById(entity repositories.IBilling) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.RemoveBillingById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BL_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
