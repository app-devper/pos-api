package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/domain/repository"
	"pos/app/domain/request"
)

func UpdateCustomerById(customerEntity repository.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("customerId")
		req := request.UpdateCustomer{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId
		result, err := customerEntity.UpdateCustomerById(id, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
