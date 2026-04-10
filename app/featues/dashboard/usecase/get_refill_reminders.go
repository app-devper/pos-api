package usecase

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetRefillReminders() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Refill reminders are not implemented server-side yet.
		// Return an empty list to keep the dashboard contract stable
		// until the underlying reminder logic is added.
		ctx.JSON(http.StatusOK, []gin.H{})
	}
}
