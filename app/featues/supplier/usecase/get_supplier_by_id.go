package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
)

func GetSupplierById(supplierEntity repositories.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("supplierId")
		result, err := supplierEntity.GetSupplierById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
