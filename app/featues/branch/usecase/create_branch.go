package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreateBranch(
	branchEntity repositories.IBranch,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Branch{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).Error("bind create branch request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		req.CreatedBy = userId

		sequence, err := sequenceEntity.NextSequence(constant.BRANCH)
		if err != nil {
			logrus.WithError(err).WithField("createdBy", userId).Error("get branch sequence failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, err.Error())
			return
		}
		if sequence == nil {
			logrus.WithField("createdBy", userId).Error("branch sequence not available")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, "branch sequence not available")
			return
		}
		req.Code = sequence.GenerateCode()

		result, err := branchEntity.CreateBranch(req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"code":      req.Code,
				"createdBy": req.CreatedBy,
			}).Error("create branch failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BR_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
