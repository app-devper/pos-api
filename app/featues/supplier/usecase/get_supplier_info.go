package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func GetSupplierInfo(supplierEntity repository.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientId := ctx.GetString("ClientId")
		result, err := supplierEntity.GetSupplierByClientId(clientId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
