package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/domain/repository"
	"pos/app/featues/request"
)

func CreateOrder(orderEntity repository.IOrder, productEntity repository.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Order{}
		if err := ctx.ShouldBind(&req); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		totalCost := 0.0
		for index, item := range req.Items {
			req.Items[index].CostPrice = productEntity.GetTotalCostPrice(item.ProductId, item.Quantity)
			totalCost += req.Items[index].CostPrice
		}
		req.TotalCost = totalCost
		userId := ctx.GetString("UserId")
		req.CreatedBy = userId

		result, err := orderEntity.CreateOrder(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, item := range req.Items {
			_, _ = productEntity.RemoveQuantityById(item.ProductId, item.Quantity)
		}

		if req.Message != "" {
			date := utils.ToFormat(result.CreatedDate)
			_, _ = utils.NotifyMassage("รายการวันที่ " + date + "\n\n" + req.Message)
		}

		ctx.JSON(http.StatusOK, result)
	}
}
