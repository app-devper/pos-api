package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
)

func DeleteCustomerById(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerId := ctx.Param("customerId")
		result, err := customerEntity.RemoveCustomerById(customerId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
