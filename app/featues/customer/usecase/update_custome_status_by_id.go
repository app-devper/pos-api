package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repository"
	"pos/app/domain/request"
)

func UpdateCustomerStatusById(customerEntity repository.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("customerId")
		req := request.UpdateCustomerStatus{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId
		result, err := customerEntity.UpdateCustomerStatusById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
