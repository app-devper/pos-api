package usecase

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/domain/repository"
	"pos/app/domain/request"
	"time"
)

func GetProductLogsExpireNotify(productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		location := utils.GetLocation()
		today := time.Now().In(location)
		startDate := utils.Bod(today)
		endDate := startDate.Add(24 * time.Hour)
		req := request.GetExpireRange{
			StartDate: startDate,
			EndDate:   endDate,
		}
		result, err := productEntity.GetProductLotsExpire(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if len(result) > 0 {
			var message = ""
			var no = 1
			for _, item := range result {
				message += fmt.Sprintf("%d. %s Lot: %s\n", no, item.Product.Name, item.LotNumber)
				no += 1
			}

			date := utils.ToFormatDate(today)
			_, _ = utils.NotifyMassage("สินค้าหมดอายุวันที่ " + date + "\n\n" + message)
		}

		ctx.JSON(http.StatusOK, result)
	}
}