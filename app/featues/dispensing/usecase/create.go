package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateDispensingLog(entity repositories.IDispensingLog) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.DispensingLog{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DI_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = ctx.GetString("BranchId")
		result, err := entity.CreateDispensingLog(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DI_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
