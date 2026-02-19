package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

func GetProductLotsExpireNotify(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		location := utils.GetLocation()
		today := time.Now().In(location)
		startDate := utils.Bod(today)
		endDate := startDate.Add(24 * time.Hour)
		req := request.GetProductLotsExpireRange{
			StartDate: startDate.UTC(),
			EndDate:   endDate.UTC(),
		}
		result, err := productEntity.GetProductLotsExpireNotify(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "success",
			"data":    result,
		})

	}
}
