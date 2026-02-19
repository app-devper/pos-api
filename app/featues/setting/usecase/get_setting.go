package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetSetting(settingEntity repositories.ISetting) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := settingEntity.GetSettingByBranchId(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.SE_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
