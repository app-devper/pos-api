package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func DeleteEmployeeById(employeeEntity repositories.IEmployee) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		employeeId := ctx.Param("employeeId")
		result, err := employeeEntity.RemoveEmployeeById(employeeId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.EM_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
