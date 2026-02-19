package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetCreditNotes(entity repositories.ICreditNote) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetCreditNotes(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CN_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetCreditNoteById(entity repositories.ICreditNote) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetCreditNoteById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CN_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdateCreditNoteById(entity repositories.ICreditNote) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdateCreditNote{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CN_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdateCreditNoteById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CN_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeleteCreditNoteById(entity repositories.ICreditNote) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.RemoveCreditNoteById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.CN_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
