package usecase

import (
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func GetRefillReminders(dispensingLogEntity repositories.IDispensingLog) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		branchId := ctx.GetString("BranchId")
		result, err := dispensingLogEntity.GetRefillReminders(branchId, 30)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"branchId": branchId,
				"days":     30,
			}).Error("get refill reminders failed")
			errcode.Abort(ctx, http.StatusBadRequest, errcode.DA_BAD_REQUEST_002, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
