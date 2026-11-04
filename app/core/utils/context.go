package utils

import (
	"context"
	"time"
)

// ContextTimeout bounds one repository call.
const ContextTimeout = 10 * time.Second

func InitContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), ContextTimeout)
	return ctx, cancel
}
