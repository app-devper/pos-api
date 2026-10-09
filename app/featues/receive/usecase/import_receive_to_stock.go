package usecase

import (
	"context"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/core/utils"
	"pos/app/data/entities"
	"pos/app/data/repositories"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ImportReceiveToStock moves a Receive into Stock. Oversell settlement happens
// inside the same transaction, in the Stock ledger (ADR-0002).
// receiveImporter imports a Receive through the Stock ledger.
type receiveImporter interface {
	ImportReceive(ctx context.Context, receiveID, branchID, by string) (*entities.Receive, error)
}

func ImportReceiveToStock(receiveEntity repositories.IReceive, ledger receiveImporter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		receiveId := ctx.Param("receiveId")

		userId := utils.GetUserId(ctx)
		branchId := utils.GetBranchId(ctx)
		if _, err := receiveEntity.GetReceiveById(receiveId, branchId); err != nil {
			errcode.AbortLookup(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}
		result, err := ledger.ImportReceive(ctx.Request.Context(), receiveId, branchId, userId)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"receiveId": receiveId, "userId": userId, "branchId": branchId,
			}).Error("import receive to stock failed")
			errcode.AbortLedger(ctx, err, errcode.RC_BAD_REQUEST_002)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}
