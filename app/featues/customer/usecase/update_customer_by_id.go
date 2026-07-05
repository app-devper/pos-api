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

func UpdateCustomerById(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("customerId")
		req := request.UpdateCustomer{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).WithField("customerId", id).Error("bind update customer request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_001, err.Error())
			return
		}
		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId
		result, err := customerEntity.UpdateCustomerById(id, req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"customerId": id,
				"updatedBy":  req.UpdatedBy,
			}).Error("update customer failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
