package utils

import "github.com/gin-gonic/gin"

func GetUserId(ctx *gin.Context) string {
	userId := ctx.GetString("UserId")
	return userId
}
