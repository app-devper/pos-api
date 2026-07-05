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

func GetPatients(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPatients(branchId)
		if err != nil {
			logrus.WithError(err).WithField("branchId", branchId).Error("get patients failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetPatientById(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPatientById(id, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"id":       id,
				"branchId": branchId,
			}).Error("get patient by id failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetPatientByCustomerCode(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code := ctx.Param("customerCode")
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPatientByCustomerCode(code, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"customerCode": code,
				"branchId":     branchId,
			}).Error("get patient by customer code failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdatePatientById(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdatePatient{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).WithField("id", id).Error("bind update patient request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		branchId := ctx.GetString("BranchId")
		result, err := entity.UpdatePatientById(id, req, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"id":        id,
				"branchId":  branchId,
				"updatedBy": req.UpdatedBy,
			}).Error("update patient failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeletePatientById(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		branchId := ctx.GetString("BranchId")
		userId := utils.GetUserId(ctx)
		result, err := entity.RemovePatientById(id, branchId, userId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"id":       id,
				"branchId": branchId,
				"userId":   userId,
			}).Error("delete patient failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
