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

func CreatePatient(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Patient{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).Error("bind create patient request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = ctx.GetString("BranchId")
		result, err := entity.CreatePatient(req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":  req.BranchId,
				"createdBy": req.CreatedBy,
			}).Error("create patient failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
