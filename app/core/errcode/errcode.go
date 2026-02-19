package errcode

import "github.com/gin-gonic/gin"

type AppError struct {
	ErrCode string `json:"errcode"`
	Error   string `json:"error"`
}

func Abort(ctx *gin.Context, httpStatus int, code string, msg string) {
	ctx.AbortWithStatusJSON(httpStatus, AppError{ErrCode: code, Error: msg})
}
