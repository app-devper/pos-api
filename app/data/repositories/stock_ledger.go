package repositories

import (
	"context"
	"sync"

	"pos/app/data/ledger"
	"pos/db"
)

var ledgers = struct {
	sync.Mutex
	byResource map[*db.Resource]*ledger.Ledger
}{byResource: map[*db.Resource]*ledger.Ledger{}}

// newLedger is the one Stock ledger (ADR-0001) over this database, shared by
// every repository that records through it. Its codes follow this
// database's document sequences, the same series as before.
func newLedger(resource *db.Resource) *ledger.Ledger {
	ledgers.Lock()
	defer ledgers.Unlock()
	if l, ok := ledgers.byResource[resource]; ok {
		return l
	}
	sequences := &sequenceEntity{sequenceRepo: resource.PosDb.Collection("sequences")}
	l := ledger.New(resource.Client, resource.PosDb, func(ctx context.Context, field, prefix string) (string, error) {
		sequence, err := sequences.nextSequenceWithContext(ctx, field)
		if err != nil {
			return "", err
		}
		return prefix + sequence.GenerateCode(), nil
	})
	ledgers.byResource[resource] = l
	return l
}
