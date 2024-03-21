package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
)

func GetCategoryById(entity repositories.ICategory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		categoryId := ctx.Param("categoryId")
		result, err := entity.GetCategoryById(categoryId)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
