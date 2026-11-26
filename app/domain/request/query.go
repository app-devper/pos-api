package request

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// BoolQuery reads the boolean query parameter name: absent is false; "true"
// or "false" in any case is that value; anything else is not ok, and the
// handler answers 400.
func BoolQuery(ctx *gin.Context, name string) (value bool, ok bool) {
	raw, present := ctx.GetQuery(name)
	if !present {
		return false, true
	}
	switch strings.ToLower(raw) {
	case strconv.FormatBool(true):
		return true, true
	case strconv.FormatBool(false):
		return false, true
	default:
		return false, false
	}
}
