package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repository"
)

func GetCategories(entity repository.ICategory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := entity.GetCategoryAll()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
