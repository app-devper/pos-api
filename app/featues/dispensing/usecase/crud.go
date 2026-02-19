package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
)

func GetDispensingLogs(entity repositories.IDispensingLog) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := entity.GetDispensingLogs(branchId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DI_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetDispensingLogById(entity repositories.IDispensingLog) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		result, err := entity.GetDispensingLogById(id)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DI_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

func GetDispensingLogsByPatientId(entity repositories.IDispensingLog) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		patientId := ctx.Param("patientId")
		result, err := entity.GetDispensingLogsByPatientId(patientId)
		if err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DI_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
