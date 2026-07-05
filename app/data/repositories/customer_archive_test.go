package repositories

import (
	"testing"

	"pos/app/domain/constant"
)

func TestCustomerArchiveStatusConstantAvailable(t *testing.T) {
	if constant.ARCHIVED != "ARCHIVED" {
		t.Fatalf("expected archived status constant, got %s", constant.ARCHIVED)
	}
}
