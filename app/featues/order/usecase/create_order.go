package usecase

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
)

func CreateOrder(
	orderEntity repositories.IOrder,
	productEntity repositories.IProduct,
	sequenceEntity repositories.ISequence,
) gin.HandlerFunc {
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

		userId := utils.GetUserId(ctx)
		req.CreatedBy = userId

		sequence, _ := sequenceEntity.NextSequence(constant.ORDER)
		if sequence != nil {
			req.Code = sequence.GenerateCode()
		}

		result, err := orderEntity.CreateOrder(req)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, item := range req.Items {
			_, _ = productEntity.RemoveQuantityById(item.ProductId, item.Quantity)
			if item.StockId != "" {
				_, _ = productEntity.RemoveProductStockQuantityById(item.StockId, item.Quantity)

				// Add product history
				unit, _ := productEntity.GetProductUnitById(item.UnitId)
				balance := productEntity.GetProductStockBalance(item.ProductId, unit.Id.Hex())
				_, _ = productEntity.CreateProductHistory(request.AddOrderItemProductHistory(item.ProductId, unit.Unit, item, balance, req.CreatedBy))
			} else {
				//
			}
		}

		if req.Message != "" {
			date := utils.ToFormat(result.CreatedDate)
			_, _ = utils.NotifyMassage("รายการวันที่ " + date + "\n\n" + req.Message)
		}

		ctx.JSON(http.StatusOK, result)
	}
}
