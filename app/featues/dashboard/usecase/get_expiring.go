package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

func GetExpiringProducts(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		now := time.Now()
		sixMonthsLater := now.AddDate(0, 6, 0)
		param := request.GetProductLotsExpireRange{
			StartDate: now,
			EndDate:   sixMonthsLater,
		}
		result, err := productEntity.GetProductLotsExpireNotify(param)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
