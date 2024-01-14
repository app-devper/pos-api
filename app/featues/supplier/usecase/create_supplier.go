package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/domain/repository"
	"pos/app/domain/request"
)

func CreateSupplier(supplierEntity repository.ISupplier) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Supplier{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userId := utils.GetUserId(ctx)
		clientId := ctx.GetString("ClientId")
		req.CreatedBy = userId

		result, err := supplierEntity.CreateSupplierByClientId(clientId, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
