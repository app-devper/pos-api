package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetPromotions(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPromotions(branchId)
		if err != nil {
			logrus.WithError(err).WithField("branchId", branchId).Error("get promotions failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetPromotionById(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPromotionById(id, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"id":       id,
				"branchId": branchId,
			}).Error("get promotion by id failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdatePromotionById(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		branchId := ctx.GetString("BranchId")
		req := request.UpdatePromotion{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).WithField("id", id).Error("bind update promotion request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdatePromotionById(id, branchId, req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"id":        id,
				"branchId":  branchId,
				"updatedBy": req.UpdatedBy,
			}).Error("update promotion failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeletePromotionById(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		branchId := ctx.GetString("BranchId")
		userId := utils.GetUserId(ctx)
		result, err := entity.RemovePromotionById(id, branchId, userId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"id":       id,
				"branchId": branchId,
				"userId":   userId,
			}).Error("delete promotion failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
