package usecase

import (
	"errors"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/promotion"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

func ApplyPromotion(entity repositories.IPromotion) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ApplyPromotion{}
		if err := ctx.ShouldBind(&req); err != nil {
			logrus.WithError(err).Error("bind apply promotion request failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_001, err.Error())
			return
		}
		branchId := ctx.GetString("BranchId")

		promo, err := entity.GetPromotionByCode(req.PromotionCode, branchId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":      branchId,
				"promotionCode": req.PromotionCode,
			}).Error("get promotion by code failed")
			if errors.Is(err, mongo.ErrNoDocuments) {
				// A code the cashier typed that names no live promotion is a
				// rule of the domain, not a missing resource.
				errcode.Abort(ctx, http.StatusBadRequest, errcode.PM_BAD_REQUEST_002, "promotion not found or expired")
				return
			}
			errcode.AbortLookup(ctx, err, errcode.PM_BAD_REQUEST_002)
			return
		}

		productIds := make([]string, len(promo.ProductIds))
		for i, id := range promo.ProductIds {
			productIds[i] = id.Hex()
		}
		terms := promotion.Terms{Type: promo.Type, Value: promo.Value, MinPurchase: promo.MinPurchase,
			MaxDiscount: promo.MaxDiscount, ProductIds: productIds}
		discount, err := terms.Discount(req.OrderTotal, req.ProductIds)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId":      branchId,
				"promotionCode": req.PromotionCode,
				"orderTotal":    req.OrderTotal,
			}).Warn("apply promotion rejected")
			errcode.Reject(ctx, errcode.PM_BAD_REQUEST_002, err)
			return
		}

		result := request.ApplyPromotionResult{
			PromotionId: promo.Id.Hex(),
			Code:        promo.Code,
			Name:        promo.Name,
			Type:        promo.Type,
			Discount:    discount,
		}
		ctx.JSON(http.StatusOK, result)
	}
}
