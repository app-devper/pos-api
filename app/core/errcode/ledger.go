package errcode

import (
	"errors"
	"net/http"

	"pos/app/data/ledger"

	"github.com/gin-gonic/gin"
)

// AbortLedger answers a Stock ledger write that failed, so the caller can
// tell what to do next:
//
//   - the event breaks a rule (ledger.Rejected): 400 with its reason;
//   - a document it names does not exist (ledger.ErrNotFound): 404;
//   - Stock or a Line changed under it (ledger.ErrConflict): 409, and
//     resending the same request is safe;
//   - anything else: 400 with the feature's code, as before.
func AbortLedger(ctx *gin.Context, err error, code string) {
	var rejected *ledger.Rejected
	switch {
	case errors.As(err, &rejected):
		Abort(ctx, http.StatusBadRequest, code, rejected.Reason)
	case errors.Is(err, ledger.ErrNotFound):
		Abort(ctx, http.StatusNotFound, SY_NOT_FOUND_002, err.Error())
	case errors.Is(err, ledger.ErrConflict):
		Abort(ctx, http.StatusConflict, SY_CONFLICT_001, err.Error())
	default:
		Abort(ctx, http.StatusBadRequest, code, err.Error())
	}
}
