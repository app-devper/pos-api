package errcode

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type AppError struct {
	ErrCode string `json:"errcode"`
	Error   string `json:"error"`
}

func Abort(ctx *gin.Context, httpStatus int, code string, msg string) {
	ctx.AbortWithStatusJSON(httpStatus, AppError{ErrCode: code, Error: msg})
}

// AbortLookup answers a failed read of one document by id. A document that
// does not exist and one that belongs to another branch are the same answer,
// 404: the repository read it with the caller's branch in the query. Any
// other failure is 400 with code.
func AbortLookup(ctx *gin.Context, err error, code string) {
	if errors.Is(err, mongo.ErrNoDocuments) {
		Abort(ctx, http.StatusNotFound, SY_NOT_FOUND_002, "not found in this branch")
		return
	}
	Abort(ctx, http.StatusBadRequest, code, err.Error())
}
