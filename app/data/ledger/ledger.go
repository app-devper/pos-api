// Package ledger is the only way a Stock's quantity changes (ADR-0001).
//
// Each method records one domain event in one transaction: the quantity
// change, Oversell settlement for every Stock that rises (ADR-0002), and
// exactly one product history row per Stock touched, carrying the balance its
// branch, Product and Unit hold once the event is done. Callers say what
// happened; they never see a session, a history row or a settlement.
package ledger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// Codes issues the next document code for a sequence field, inside the
// caller's transaction (e.g. "AJ-" for Adjustments).
type Codes func(ctx context.Context, field, prefix string) (string, error)

type Ledger struct {
	client *mongo.Client
	db     *mongo.Database
	codes  Codes
}

func New(client *mongo.Client, db *mongo.Database, codes Codes) *Ledger {
	return &Ledger{client: client, db: db, codes: codes}
}

// Rejected is a business rule the event broke. Reason is shown to the user as
// is and the event is never worth retrying.
type Rejected struct{ Reason string }

func (r *Rejected) Error() string { return r.Reason }

func reject(format string, a ...any) error { return &Rejected{Reason: fmt.Sprintf(format, a...)} }

// ErrNotFound: a document the event names does not exist, or not in the
// caller's branch.
var ErrNotFound = errors.New("not found")

// ErrConflict: Stock or a Line changed under the event. Resending is safe.
var ErrConflict = errors.New("stock changed while recording; try again")

// timeout bounds one event, retries included, as utils.InitContext does.
const timeout = 10 * time.Second

// missing turns "no such document" into ErrNotFound and passes every other
// error through untouched, so transient errors stay retryable.
func missing(err error, what string, a ...any) error {
	if errors.Is(err, mongo.ErrNoDocuments) {
		return fmt.Errorf("%w: "+what, append([]any{ErrNotFound}, a...)...)
	}
	return err
}

// run executes fn as one snapshot/majority transaction and writes the history
// rows for every Stock it touched before committing. fn may run more than once
// (transient transaction errors are retried), so it must only write through
// the book it is given.
func (l *Ledger) run(parent context.Context, fn func(b *book) (any, error)) (any, error) {
	if l.client == nil {
		return nil, errors.New("stock ledger needs a Mongo client (transactions)")
	}
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	session, err := l.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)
	return session.WithTransaction(ctx, func(sc mongo.SessionContext) (any, error) {
		b := newBook(l, sc)
		result, err := fn(b)
		if err != nil {
			return nil, err
		}
		if err := b.flush(); err != nil {
			return nil, err
		}
		return result, nil
	}, options.Transaction().SetReadConcern(readconcern.Snapshot()).SetWriteConcern(writeconcern.Majority()))
}
