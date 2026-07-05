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

func CreatePromotion(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Promotion{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).Error("bind create promotion request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = ctx.GetString("BranchId")
		result, err := entity.CreatePromotion(req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":  req.BranchId,
				"createdBy": req.CreatedBy,
			}).Error("create promotion failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
