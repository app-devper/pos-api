package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetQuotations(entity repositories.IQuotation) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetQuotations(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.QT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetQuotationById(entity repositories.IQuotation) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetQuotationById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.QT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateQuotationById(entity repositories.IQuotation) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateQuotation{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.QT_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdateQuotationById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.QT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteQuotationById(entity repositories.IQuotation) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.RemoveQuotationById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.QT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
