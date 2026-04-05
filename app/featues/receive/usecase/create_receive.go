package usecase

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/constant"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreateReceive(receiveEntity repositories.IReceive, sequenceEntity repositories.ISequence, productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.Receive{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_001, err.Error())
			return
		}

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)

		sequence, err := sequenceEntity.NextSequence(constant.RECEIVE)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}
		if sequence == nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, "receive sequence not available")
			return
		}
		req.Code = sequence.GenerateCode()
		req.UpdatedBy = userId
		req.BranchId = branchId

		result, err := receiveEntity.CreateReceive(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, err.Error())
			return
		}

		receiveId := result.Id.Hex()
		rollback := func(cause error) {
			if _, rollbackErr := receiveEntity.RemoveReceiveById(receiveId); rollbackErr != nil {
				logrus.WithError(rollbackErr).WithFields(logrus.Fields{
					"receiveId": receiveId,
					"branchId":  branchId,
					"code":      req.Code,
				}).Error("create receive rollback failed")
			}
			logrus.WithError(cause).WithFields(logrus.Fields{
				"receiveId": receiveId,
				"branchId":  branchId,
				"code":      req.Code,
			}).Error("create receive rolled back")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.RC_BAD_REQUEST_002, cause.Error())
		}
		var totalCost float64

		for _, item := range req.Items {
			if item.ProductId == "" || item.Quantity <= 0 {
				continue
			}

			product, pErr := productEntity.GetProductById(item.ProductId)
			if pErr != nil {
				rollback(fmt.Errorf("failed to load product %s: %w", item.ProductId, pErr))
				return
			}
			if product == nil {
				rollback(fmt.Errorf("product %s not found", item.ProductId))
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
				ReceiveId:    receiveId,
				ReceiveCode:  req.Code,
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

			if _, err := receiveEntity.CreateReceiveItem(receiveId, "", item.ProductId, productReq); err != nil {
				rollback(fmt.Errorf("failed to create receive item for product %s: %w", item.ProductId, err))
				return
			}

			totalCost += item.CostPrice * float64(item.Quantity)
		}

		if totalCost > 0 {
			if _, err := receiveEntity.UpdateReceiveTotalCostById(receiveId, totalCost); err != nil {
				rollback(fmt.Errorf("failed to update receive total cost: %w", err))
				return
			}
		}

		ctx.JSON(http.StatusOK, result)
	}
}
