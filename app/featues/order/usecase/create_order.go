package usecase

import (
	"net/http"
	"pos/app/core/errcode"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// CreateOrder records a Sale. Every till sends its Sale with a saleId and
// pos-api prices it and draws its Stock (ADR-0001). A request without one
// comes from a till that priced the Sale itself, and is refused.
func CreateOrder(sales seller) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var probe struct {
			SaleId string `json:"saleId"`
		}
		if err := ctx.ShouldBindBodyWith(&probe, binding.JSON); err != nil {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001, err.Error())
			return
		}
		if probe.SaleId == "" {
			errcode.Abort(ctx, http.StatusBadRequest, errcode.OR_BAD_REQUEST_001,
				"saleId is required: this till is out of date, reload it")
			return
		}
		recordSale(ctx, sales)
	}
}
