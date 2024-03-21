package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/data/repositories"
	"pos/app/domain/request"
)

func UpdateCategoryById(entity repositories.ICategory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		categoryId := ctx.Param("categoryId")
		req := request.Category{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := entity.UpdateCategoryById(categoryId, req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
