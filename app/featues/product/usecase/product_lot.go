package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

func GetAllLots(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.GetProductLotsExpireRange{}
		// Date range is optional for listing all lots
		_ = ctx.ShouldBindQuery(&req)
		result, err := productEntity.GetProductLots(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetLotById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lotId := ctx.Param("lotId")
		result, err := productEntity.GetProductLotById(lotId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func CreateLot(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		req := request.ProductLot{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := productEntity.CreateProductLot(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateLotById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lotId := ctx.Param("lotId")
		req := request.UpdateProductLot{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := productEntity.UpdateProductLotById(lotId, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteLotById(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		lotId := ctx.Param("lotId")
		result, err := productEntity.RemoveProductLotById(lotId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetProductLotsExpireNotify(productEntity repositories.IProduct) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		location := utils.GetLocation()
		today := time.Now().In(location)
		startDate := utils.Bod(today)
		endDate := startDate.Add(24 * time.Hour)
		req := request.GetProductLotsExpireRange{
			StartDate: startDate.UTC(),
			EndDate:   endDate.UTC(),
		}
		result, err := productEntity.GetProductLotsExpireNotify(req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PD_BAD_REQUEST_002, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":  http.StatusOK,
			"message": "success",
			"data":    result,
		})

	}
}
