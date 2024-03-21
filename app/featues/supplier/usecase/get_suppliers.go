package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
)

func GetSuppliers(supplierEntity repositories.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := supplierEntity.GetSuppliers()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
