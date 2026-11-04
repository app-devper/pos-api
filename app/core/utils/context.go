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

// DetachedFrom bounds one repository call that must finish once it has
// started: it keeps the parent's values but not its cancellation, so a client
// that disconnects does not tear the call apart. It still ends after
// ContextTimeout.
func DetachedFrom(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(parent), ContextTimeout)
}

// InitContext bounds one repository call that has no request to belong to.
func InitContext() (context.Context, context.CancelFunc) {
	return ContextFrom(context.Background())
}
