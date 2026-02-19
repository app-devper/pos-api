package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetSupplierInfo(supplierEntity repositories.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientId := ctx.GetString("ClientId")
		result, err := supplierEntity.GetSupplierByClientId(clientId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SU_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
