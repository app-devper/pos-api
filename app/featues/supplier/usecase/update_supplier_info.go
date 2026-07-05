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

func UpdateSupplierInfo(supplierEntity repositories.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Supplier{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).Error("bind update supplier info request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SU_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		clientId := ctx.GetString("ClientId")
		req.UpdatedBy = userId

		result, err := supplierEntity.CreateOrUpdateSupplierByClientId(clientId, req)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"clientId":  clientId,
				"updatedBy": req.UpdatedBy,
			}).Error("update supplier info failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SU_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
