package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func CreateBilling(entity repositories.IBilling, sequenceEntity repositories.ISequence) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Billing{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BL_BAD_REQUEST_001, err.Error())
			return
		}
		req.CreatedBy = utils.GetUserId(ctx)
		req.BranchId = ctx.GetString("BranchId")
		sequence, _ := sequenceEntity.NextSequence(constant.BILLING)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}
		result, err := entity.CreateBilling(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.BL_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
