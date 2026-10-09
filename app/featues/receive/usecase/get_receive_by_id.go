package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetReceiveById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("receiveId")
		result, err := receiveEntity.GetReceiveById(id, utils.GetBranchId(ctx))
		if err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}
		items, err := receiveEntity.GetReceiveItemsByReceiveId(id)
		if err == nil && len(items) > 0 {
			result.Items = items
		}
		ctx.JSON(http.StatusOK, result)
	}
}
