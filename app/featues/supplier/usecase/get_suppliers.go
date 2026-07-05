package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetSuppliers(supplierEntity repositories.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := supplierEntity.GetSuppliers()
		if err != nil {
			logrus.WithError(err).Error("get suppliers failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SU_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
