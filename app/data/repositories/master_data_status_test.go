package repositories

import (
	"testing"

	"pos/app/domain/constant"
)

func TestMasterDataInactiveStatusAvailable(t *testing.T) {
	if constant.INACTIVE != "INACTIVE" {
		t.Fatalf("expected inactive status constant, got %s", constant.INACTIVE)
	}
}
