package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repository"
	"pos/app/domain/request"
)

func GetReceivesRange(receiveEntity repository.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetReceiveRange{}
		if err := ctx.ShouldBindQuery(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := receiveEntity.GetReceives(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
