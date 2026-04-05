package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func CreateProductUnit(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.CreateProductUnit{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId
		unit, err := productEntity.CreateProductUnitCascade(req, ctx.GetString("BranchId"), userId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, unit)
	}
}

func GetProductUnitsByProductId(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		productId := ctx.Param("productId")
		result, err := productEntity.GetProductUnitsByProductId(productId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateProductUnitById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductUnit{}
		id := ctx.Param("unitId")
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		userId := ctx.GetString("UserId")
		req.UpdatedBy = userId

		unit, err := productEntity.UpdateProductUnitById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		appendProductUnitHistory(ctx.GetString("BranchId"), productEntity, func(branchId string) request.ProductHistory {
			updUnitHistory := request.UpdateProductUnitHistory(req.ProductId, req)
			updUnitHistory.BranchId = branchId
			return updUnitHistory
		})

		ctx.JSON(http.StatusOK, unit)
	}
}

func RemoveProductUnitById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("unitId")
		userId := ctx.GetString("UserId")
		result, err := productEntity.RemoveProductUnitCascade(id, ctx.GetString("BranchId"), userId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func appendProductUnitHistory(branchId string, productEntity repositories.IProduct, buildHistory func(branchId string) request.ProductHistory) {
	history := buildHistory(branchId)
	if _, err := productEntity.CreateProductHistory(history); err != nil {
		logrus.WithError(err).WithField("branchId", branchId).Error("failed to create product unit history")
	}
}
