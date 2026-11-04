package utils

import (
	"context"
	"time"
)

// ContextTimeout bounds one repository call.
const ContextTimeout = 10 * time.Second

// ContextFrom bounds one repository call that runs under its caller's
// context: it ends when the request ends or after ContextTimeout, whichever
// comes first.
func ContextFrom(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, ContextTimeout)
}

// InitContext bounds one repository call that has no request to belong to.
func InitContext() (context.Context, context.CancelFunc) {
	return ContextFrom(context.Background())
}
