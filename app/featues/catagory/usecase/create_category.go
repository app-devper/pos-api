package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateCategory(entity repositories.ICategory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Category{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CA_BAD_REQUEST_001, err.Error())
			return
		}
		result, err := entity.CreateCategory(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CA_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
