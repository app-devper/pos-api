package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func UpdateDefaultCategoryById(entity repositories.ICategory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		categoryId := ctx.Param("categoryId")
		result, err := entity.UpdateDefaultCategoryById(categoryId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
