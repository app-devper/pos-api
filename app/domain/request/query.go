package request

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// BoolQuery reads the boolean query parameter name: absent is false; "true"
// or "false" in any case is that value; anything else is not ok, and the
// handler answers 400. A parameter given more than once is not ok either.
func BoolQuery(ctx *gin.Context, name string) (value bool, ok bool) {
	raw := ctx.QueryArray(name)
	if len(raw) == 0 {
		return false, true
	}
	if len(raw) > 1 {
		return false, false
	}
	switch strings.ToLower(raw[0]) {
	case strconv.FormatBool(true):
		return true, true
	case strconv.FormatBool(false):
		return false, true
	default:
		return false, false
	}
}
