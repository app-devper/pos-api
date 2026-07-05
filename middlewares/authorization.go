package middlewares

import (
	"net/http"
	"pos/app/core/errcode"

	"github.com/gin-gonic/gin"
)

func RequireAuthorization(auths ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetString("Role")
		if role == "" {
			invalidRequest(ctx)
			return
		}
		for _, auth := range auths {
			if role == auth {
				ctx.Next()
				return
			}
		}
		notPermission(ctx)
	}
}

func invalidRequest(ctx *gin.Context) {
	errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_001, "Invalid request, restricted endpoint")
}

func notPermission(ctx *gin.Context) {
	errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "Don't have permission")
}
