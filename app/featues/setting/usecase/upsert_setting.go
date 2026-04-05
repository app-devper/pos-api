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

func UpsertSetting(settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Setting{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).Error("bind upsert setting request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SE_BAD_REQUEST_001, err.Error())
			return
		}
		req.BranchId = ctx.GetString("BranchId")
		req.UpdatedBy = utils.GetUserId(ctx)

		result, err := settingEntity.UpsertSetting(req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":  req.BranchId,
				"updatedBy": req.UpdatedBy,
			}).Error("upsert setting failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SE_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
