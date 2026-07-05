package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func UpdateCategoryById(entity repositories.ICategory) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		categoryId := ctx.Param("categoryId")
		req := request.Category{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).WithField("categoryId", categoryId).Error("bind update category request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CA_BAD_REQUEST_001, err.Error())
			return
		}
		result, err := entity.UpdateCategoryById(categoryId, req)
		if err != nil {
			logrus.WithError(err).WithField("categoryId", categoryId).Error("update category failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
