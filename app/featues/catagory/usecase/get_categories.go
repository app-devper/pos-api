package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetCategories(entity repositories.ICategory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := entity.GetCategoryAll()
		if err != nil {
			logrus.WithError(err).Error("get categories failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CA_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
