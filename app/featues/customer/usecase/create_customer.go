package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	repositories "pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateCustomer(customerEntity repositories.ICustomer, sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Customer{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_001, err.Error())
			return
		}

		if req.CustomerType == "" {
			req.CustomerType = constant.CustomerTypeGeneral
		}

		userId := utils.GetUserId(ctx)
		req.CreatedBy = userId

		sequence, _ := sequenceEntity.NextSequence(constant.MEMBER)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}

		result, err := customerEntity.CreateCustomer(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CU_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
