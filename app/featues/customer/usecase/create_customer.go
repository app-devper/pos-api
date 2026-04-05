package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	repositories "pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreateCustomer(customerEntity repositories.ICustomer, sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Customer{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).Error("bind create customer request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_001, err.Error())
			return
		}

		if req.CustomerType == "" {
			req.CustomerType = constant.CustomerTypeGeneral
		}

		userId := utils.GetUserId(ctx)
		req.CreatedBy = userId

		sequence, err := sequenceEntity.NextSequence(constant.MEMBER)
		if err != nil {
			logrus.WithError(err).WithField("createdBy", userId).Error("get customer sequence failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}
		if sequence == nil {
			logrus.WithField("createdBy", userId).Error("customer sequence not available")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, "customer sequence not available")
			return
		}
		req.Code = sequence.GenerateCode()

		result, err := customerEntity.CreateCustomer(req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"code":      req.Code,
				"createdBy": req.CreatedBy,
			}).Error("create customer failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
