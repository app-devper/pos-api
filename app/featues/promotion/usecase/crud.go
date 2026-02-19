package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetPromotions(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPromotions(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetPromotionById(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetPromotionById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdatePromotionById(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdatePromotion{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdatePromotionById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeletePromotionById(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.RemovePromotionById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
