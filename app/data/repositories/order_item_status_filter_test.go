package repositories

import (
	"testing"

	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
)

func TestConfirmedOrderItemStatusMatchClausesIncludeLegacyAndConfirmedStatuses(t *testing.T) {
	clauses := confirmedOrderItemStatusMatchClauses()
	if len(clauses) != 4 {
		t.Fatalf("expected 4 clauses, got %d", len(clauses))
	}

	if exists, ok := clauses[0]["status"].(bson.M); ok {
		if exists["$exists"] != false {
			t.Fatalf("expected first clause to match missing status, got %+v", clauses[0])
		}
	} else {
		t.Fatalf("expected first clause to be missing-status match, got %+v", clauses[0])
	}

	if clauses[1]["status"] != "" {
		t.Fatalf("expected second clause to match empty status, got %+v", clauses[1])
	}
	if clauses[2]["status"] != constant.ACTIVE {
		t.Fatalf("expected third clause to match ACTIVE, got %+v", clauses[2])
	}
	if clauses[3]["status"] != constant.CONFIRMED {
		t.Fatalf("expected fourth clause to match CONFIRMED, got %+v", clauses[3])
	}
}
