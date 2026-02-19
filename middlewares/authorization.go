package middlewares

import (
	"net/http"
	"pos/app/core/errcode"

	"github.com/gin-gonic/gin"
)

func RequireAuthorization(auths ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var roles []string
		roles = append(roles, ctx.GetString("Role"))
		if roles[0] == "" {
			invalidRequest(ctx)
			return
		}
		isAccessible := false
		if len(roles) < len(auths) || len(roles) == len(auths) {
			for _, auth := range auths {
				for _, role := range roles {
					if role == auth {
						isAccessible = true
						break
					}
				}
			}
		}
		if len(roles) > len(auths) {
			for _, role := range roles {
				for _, auth := range auths {
					if auth == role {
						isAccessible = true
						break
					}
				}
			}
		}
		if isAccessible == false {
			notPermission(ctx)
			return
		}
		ctx.Next()
	}
}

func invalidRequest(ctx *gin.Context) {
	errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_001, "Invalid request, restricted endpoint")
}

func notPermission(ctx *gin.Context) {
	errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "Don't have permission")
}
