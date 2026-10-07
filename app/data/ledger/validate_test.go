package ledger

import (
	"context"
	"errors"
	"testing"

	"pos/app/domain/request"
)

// Requests that cannot be recorded are refused before any transaction starts,
// so these need no database.
func TestRequestTransferRefusesBadBranchesWithoutTouchingTheDatabase(t *testing.T) {
	l := New(nil, nil, nil)
	same := "507f1f77bcf86cd799439011"
	for name, form := range map[string]request.StockTransfer{
		"invalid source":      {FromBranchId: "invalid", ToBranchId: same},
		"invalid destination": {FromBranchId: same, ToBranchId: "invalid"},
		"same branch":         {FromBranchId: same, ToBranchId: same},
		"invalid product": {FromBranchId: same, ToBranchId: "507f1f77bcf86cd799439012",
			Items: []request.StockTransferItem{{ProductId: "invalid", StockId: same, Quantity: 1}}},
	} {
		var rejected *Rejected
		if _, err := l.RequestTransfer(context.Background(), form); !errors.As(err, &rejected) {
			t.Errorf("%s: want Rejected, got %v", name, err)
		}
	}
}
