package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

func UpdateReceiveById(receiveEntity repositories.IReceive, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceive{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		id := ctx.Param("receiveId")

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)
		req.UpdatedBy = userId

		receive, err := receiveEntity.GetReceiveById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		var totalCost float64
		filteredItems := make([]request.ReceiveItem, 0, len(req.ReceiveItems))
		for _, item := range req.ReceiveItems {
			if item.ProductId == "" || item.Quantity <= 0 {
				continue
			}
			product, pErr := productEntity.GetProductById(item.ProductId)
			if pErr != nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, fmt.Errorf("failed to load product %s: %w", item.ProductId, pErr).Error())
				return
			}
			if product == nil {
				errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, fmt.Sprintf("product %s not found", item.ProductId))
				return
			}
			productReq := request.Product{
				Name:         product.Name,
				SerialNumber: product.SerialNumber,
				Price:        product.Price,
				CostPrice:    item.CostPrice,
				Unit:         product.Unit,
				Quantity:     item.Quantity,
				LotNumber:    item.LotNumber,
				ExpireDate:   time.Time{},
				ReceiveId:    id,
				ReceiveCode:  receive.Code,
				CreatedBy:    userId,
				BranchId:     branchId,
			}
			if item.ExpireDate != "" {
				if t, tErr := time.Parse(time.RFC3339, item.ExpireDate); tErr == nil {
					productReq.ExpireDate = t
				} else if t, tErr := time.Parse("2006-01-02", item.ExpireDate); tErr == nil {
					productReq.ExpireDate = t
				}
			}
			item.ExpireDate = productReq.ExpireDate.Format(time.RFC3339)
			filteredItems = append(filteredItems, item)
			totalCost += item.CostPrice * float64(item.Quantity)
		}

		req.ReceiveItems = filteredItems
		req.TotalCost = totalCost
		result, err := receiveEntity.UpdateReceiveById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateReceiveItemsById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceiveItems{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		receiveId := ctx.Param("receiveId")

		result, err := receiveEntity.UpdateReceiveItemsById(receiveId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateReceiveTotalCostById(receiveEntity repositories.IReceive) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceiveTotalCode{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		id := ctx.Param("receiveId")

		result, err := receiveEntity.UpdateReceiveTotalCostById(id, req.TotalCost)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}
