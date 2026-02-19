package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func UpdateEmployeeById(employeeEntity repositories.IEmployee) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		employeeId := ctx.Param("employeeId")
		req := request.UpdateEmployee{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.EM_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId

		result, err := employeeEntity.UpdateEmployeeById(employeeId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.EM_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
