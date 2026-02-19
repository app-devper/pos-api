package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetReceiveById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("receiveId")
		result, err := receiveEntity.GetReceiveById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
