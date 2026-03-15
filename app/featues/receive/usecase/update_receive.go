package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

		_ = receiveEntity.DeleteReceiveItemsByReceiveId(id)

		var totalCost float64
		for _, item := range req.ReceiveItems {
			if item.ProductId == "" || item.Quantity <= 0 {
				continue
			}
			product, pErr := productEntity.GetProductById(item.ProductId)
			if pErr != nil || product == nil {
				logrus.Warnf("UpdateReceive: product %s not found, skipping", item.ProductId)
				continue
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
			_, _ = receiveEntity.CreateReceiveItem(id, "", item.ProductId, productReq)
			totalCost += item.CostPrice * float64(item.Quantity)
		}

		req.TotalCost = totalCost
		result, err := receiveEntity.UpdateReceiveById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		items, _ := receiveEntity.GetReceiveItemsByReceiveId(id)
		if len(items) > 0 {
			result.Items = items
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
