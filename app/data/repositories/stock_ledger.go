package repositories

import (
	"context"

	"pos/app/data/ledger"
	"pos/db"
)

// newLedger builds the Stock ledger (ADR-0001) with this database's document
// sequences, so the codes it issues follow the same series as before.
func newLedger(resource *db.Resource) *ledger.Ledger {
	sequences := &sequenceEntity{sequenceRepo: resource.PosDb.Collection("sequences")}
	return ledger.New(resource.Client, resource.PosDb, func(ctx context.Context, field, prefix string) (string, error) {
		sequence, err := sequences.nextSequenceWithContext(ctx, field)
		if err != nil {
			return "", err
		}
		return prefix + sequence.GenerateCode(), nil
	})
}
