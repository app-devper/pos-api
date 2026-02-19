package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func UpdateCustomerStatusById(customerEntity repositories.ICustomer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("customerId")
		req := request.UpdateCustomerStatus{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_001, err.Error())
			return
		}
		userId := utils.GetUserId(ctx)
		req.UpdatedBy = userId
		result, err := customerEntity.UpdateCustomerStatusById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
