package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/repositories"
	"pos/app/domain/request"

	"github.com/gin-gonic/gin"
)

func GetPatients(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPatients(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetPatientById(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetPatientById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetPatientByCustomerCode(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		code := ctx.Param("customerCode")
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetPatientByCustomerCode(code, branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func UpdatePatientById(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		req := request.UpdatePatient{}
		if err := ctx.ShouldBind(&req); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_001, err.Error())
			return
		}
		req.UpdatedBy = utils.GetUserId(ctx)
		result, err := entity.UpdatePatientById(id, req)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func DeletePatientById(entity repositories.IPatient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.RemovePatientById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.PT_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
