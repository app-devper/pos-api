package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func DeleteSupplierById(supplierEntity repository.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("supplierId")
		result, err := supplierEntity.RemoveSupplierById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
