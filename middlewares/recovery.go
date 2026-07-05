package middlewares

import (
	"fmt"
	"net/http"
	"pos/app/core/errcode"

	"github.com/gin-gonic/gin"
)

func NewRecovery() gin.HandlerFunc {
	return gin.CustomRecovery(recoveryHandler)
}

func recoveryHandler(ctx *gin.Context, err interface{}) {
	errcode.Abort(ctx, http.StatusInternalServerError, errcode.SY_INTERNAL_001, fmt.Sprintf("%v", err))
}
