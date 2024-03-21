package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
)

func GetCustomers(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := customerEntity.GetCustomerAll()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
