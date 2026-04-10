package usecase

import (
	"errors"
	"fmt"
	"net/http"
	"pos/app/core/errcode"
	"pos/app/data/entities"
	"pos/app/domain/request"
	"time"

	"github.com/gin-gonic/gin"
)

func abortReceiveBranchMismatch(ctx *gin.Context) {
	errcode.Abort(ctx, http.StatusForbidden, errcode.SY_FORBIDDEN_002, "no permission")
}

func ensureReceiveBranchAccess(receive *entities.Receive, branchId string) error {
	if receive == nil {
		return errors.New("receive not found")
	}
	if receive.BranchId.Hex() != branchId {
		return errors.New("receive belongs to another branch")
	}
	return nil
}

func parseReceiveExpireDate(value string) (request.FlexibleTime, error) {
	if value == "" {
		return request.NewFlexibleTime(time.Time{}), nil
	}

	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return request.NewFlexibleTime(t), nil
	}
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return request.NewFlexibleTime(t), nil
	}

	return request.NewFlexibleTime(time.Time{}), fmt.Errorf("invalid expireDate: %s", value)
}
