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

		if _, err := receiveEntity.GetReceiveById(id, branchId); err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}
		items, err := receiveItems(productEntity, req.ReceiveItems)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		req.ReceiveItems = items
		result, err := receiveEntity.UpdateReceiveById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateReceiveItemsById(receiveEntity repositories.IReceive, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.UpdateReceiveItems{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		receiveId := ctx.Param("receiveId")
		if _, err := receiveEntity.GetReceiveById(receiveId, utils.GetBranchId(ctx)); err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		items, err := receiveItems(productEntity, req.ReceiveItems)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		req.ReceiveItems = items

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
		req := request.UpdateReceiveTotalCost{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}
		id := ctx.Param("receiveId")
		if _, err := receiveEntity.GetReceiveById(id, utils.GetBranchId(ctx)); err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}
		result, err := receiveEntity.UpdateReceiveTotalCostById(id, req.TotalCost)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, result)
	}
}

// receiveItems keeps the lines that receive something, checks each names a
// Product, and normalises its expiry date. A Receive's edit and its items
// edit take items through the same rules.
func receiveItems(productEntity repositories.IProduct, items []request.ReceiveItem) ([]request.ReceiveItem, error) {
	kept := make([]request.ReceiveItem, 0, len(items))
	for _, item := range items {
		if item.ProductId == "" || item.Quantity <= 0 {
			continue
		}
		product, err := productEntity.GetProductById(item.ProductId)
		if err != nil {
			return nil, fmt.Errorf("failed to load product %s: %w", item.ProductId, err)
		}
		if product == nil {
			return nil, fmt.Errorf("product %s not found", item.ProductId)
		}
		expireDate, err := parseReceiveExpireDate(item.ExpireDate)
		if err != nil {
			return nil, err
		}
		item.ExpireDate = expireDate.Time.Format(time.RFC3339)
		kept = append(kept, item)
	}
	return kept, nil
}
