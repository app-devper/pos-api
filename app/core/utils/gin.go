package utils

import "github.com/gin-gonic/gin"

func GetUserId(ctx *gin.Context) string {
	userId := ctx.GetString("UserId")
	return userId
}

func GetBranchId(ctx *gin.Context) string {
	branchId := ctx.GetString("BranchId")
	return branchId
}
