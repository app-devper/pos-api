package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func UpdateCustomerStatusById(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("customerId")
		req := request.UpdateCustomerStatus{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).WithField("customerId", id).Error("bind update customer status request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_001, err.Error())
			return
		}
		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId
		if !isValidCustomerStatus(req.Status) {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_001, fmt.Sprintf("invalid customer status: %s", req.Status))
			return
		}
		result, err := customerEntity.UpdateCustomerStatusById(id, req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"customerId": id,
				"updatedBy":  req.UpdatedBy,
			}).Error("update customer status failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
