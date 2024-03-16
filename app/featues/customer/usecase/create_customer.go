package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/domain/constant"
	"pos/app/domain/repository"
	"pos/app/domain/request"
)

func CreateCustomer(customerEntity repository.ICustomer, sequenceEntity repository.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Customer{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
