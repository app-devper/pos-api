package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
)

func CreateReceive(receiveEntity repositories.IReceive, sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Receive{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userId := utils.GetUserId(ctx)

		sequence, _ := sequenceEntity.NextSequence(constant.RECEIVE)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}
		req.UpdatedBy = userId

		result, err := receiveEntity.CreateReceive(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
