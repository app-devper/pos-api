package errcode

import (
	"context"
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
// other failure is 400 with code. A read that ran out of time is 504 and one
// the client abandoned is 499 (see abortContextError).
func AbortLookup(ctx *gin.Context, err error, code string) {
	if abortContextError(ctx, err) {
		return
	}
	if errors.Is(err, mongo.ErrNoDocuments) {
		Abort(ctx, http.StatusNotFound, SY_NOT_FOUND_002, "not found in this branch")
		return
	}
	Abort(ctx, http.StatusBadRequest, code, err.Error())
}

// AbortWrite answers a failed repository write (insert, update, delete) that
// is not recorded by the Stock ledger (those go through AbortLedger). A write
// that ran out of time is 504 and one the client abandoned is 499 (see
// abortContextError); a document whose key already exists is 409 with
// SY_CONFLICT_002, so the caller knows resending the same request will not
// help; anything else is 400 with code.
func AbortWrite(ctx *gin.Context, err error, code string) {
	if abortContextError(ctx, err) {
		return
	}
	if mongo.IsDuplicateKeyError(err) {
		Abort(ctx, http.StatusConflict, SY_CONFLICT_002, "already exists")
		return
	}
	Abort(ctx, http.StatusBadRequest, code, err.Error())
}

// statusClientClosedRequest is the de-facto (nginx) status for a client that
// left before the answer; net/http has no constant for it.
const statusClientClosedRequest = 499

// abortContextError answers a failure caused by the request's context and
// reports whether it did: context.DeadlineExceeded is 504 with
// SY_TIMEOUT_001; context.Canceled is 499 with no body, since nobody is
// listening. Any other error is left to the caller.
func abortContextError(ctx *gin.Context, err error) bool {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		Abort(ctx, http.StatusGatewayTimeout, SY_TIMEOUT_001, "timed out")
		return true
	case errors.Is(err, context.Canceled):
		ctx.AbortWithStatus(statusClientClosedRequest)
		return true
	}
	return false
}
