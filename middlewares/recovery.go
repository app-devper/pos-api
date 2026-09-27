package middlewares

import (
	"net/http"
	"pos/app/core/errcode"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(recoveryHandler)
}

// recoveryHandler logs the panic and returns a generic message, so internal
// details never reach the client.
func recoveryHandler(ctx *gin.Context, err interface{}) {
	logrus.Error("panic recovered: ", err)
	errcode.Abort(ctx, http.StatusInternalServerError, errcode.SY_INTERNAL_001, "internal server error")
}
