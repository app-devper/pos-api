package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
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

		// Add product history
		updUnitHistory := request.UpdateProductUnitHistory(req.ProductId, req)
		updUnitHistory.BranchId = ctx.GetString("BranchId")
		_, _ = productEntity.CreateProductHistory(updUnitHistory)

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
