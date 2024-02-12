package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/domain/repository"
)

func DeleteReceiveById(receiveEntity repository.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("receiveId")
		result, err := receiveEntity.RemoveReceiveById(id)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
