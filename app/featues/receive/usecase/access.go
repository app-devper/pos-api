package usecase

import (
	"fmt"
	"pos/app/domain/request"
	"time"
)

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
