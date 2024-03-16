package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repository"
	"pos/app/domain/constant"
)

func GenerateSerialNumber(sequenceEntity repository.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := sequenceEntity.NextSequence(constant.PRODUCT)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"serialNumber": result.GenerateCode()})
	}
}
