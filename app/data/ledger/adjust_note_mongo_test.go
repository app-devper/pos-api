package ledger

import (
	"context"
	"strings"
	"testing"

	"pos/app/domain/constant"
	"pos/app/domain/request"

	"go.mongodb.org/mongo-driver/bson"
)

func (f *fixture) adjustWithNote(delta int, note string) string {
	f.t.Helper()
	stock := f.stock(f.tab, 1, 0)
	req := f.adjust(stock, delta)
	req.Reason = constant.AdjustmentReasonOther
	req.Note = note
	if _, err := f.ledger.Adjust(context.Background(), req); err != nil {
		f.t.Fatal(err)
	}
	h := f.histories()
	if len(h) != 1 {
		f.t.Fatalf("want one history row, got %+v", h)
	}
	return h[0].Description
}

// An Adjustment's note reaches the stock card after the quantity and its Unit;
// without a note the row reads exactly as it always has.
func TestAdjustHistoryCarriesTheNoteOnOneLine(t *testing.T) {
	long := strings.Repeat("ก", 250)
	for _, c := range []struct{ name, note, want string }{
		{"note", "นับใหม่", "ปรับสต็อกสินค้า (อื่นๆ) จำนวน 2 TAB — นับใหม่"},
		{"no note", "", "ปรับสต็อกสินค้า (อื่นๆ) จำนวน 2 TAB"},
		{"blank note", " \n\t ", "ปรับสต็อกสินค้า (อื่นๆ) จำนวน 2 TAB"},
		{"newlines collapse", "  นับ\r\nใหม่\t\tรอบเย็น ", "ปรับสต็อกสินค้า (อื่นๆ) จำนวน 2 TAB — นับ ใหม่ รอบเย็น"},
		{"long note is cut", long, "ปรับสต็อกสินค้า (อื่นๆ) จำนวน 2 TAB — " + strings.Repeat("ก", 200) + "…"},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := newFixture(t).adjustWithNote(2, c.note); got != c.want {
				t.Fatalf("history %q, want %q", got, c.want)
			}
		})
	}
}

// The settlement suffix stays beside the quantity it explains; the note ends
// the row.
func TestAdjustHistoryPutsTheNoteAfterTheSettlement(t *testing.T) {
	f := newFixture(t)
	f.owed(f.branch, f.tab, 1)
	want := "ปรับสต็อกสินค้า (อื่นๆ) จำนวน 2 TAB (ส่งของค้างลูกค้า 1 TAB) — นับใหม่"
	if got := f.adjustWithNote(2, "นับใหม่"); got != want {
		t.Fatalf("history %q, want %q", got, want)
	}
}

// A Count's Lines are Adjustments noted with the Count's code and note.
func TestCountHistoryCarriesTheCountCodeAndNote(t *testing.T) {
	for _, note := range []string{"รอบเย็น", ""} {
		f := newFixture(t)
		stock := f.stock(f.tab, 1, 5)
		got, err := f.ledger.Count(context.Background(), request.StockCount{BranchId: f.branch.Hex(), CreatedBy: "admin", Note: note,
			Items: []request.StockCountItem{{ProductId: f.product.Hex(), StockId: stock.Hex(), Counted: 3}}})
		if err != nil {
			t.Fatal(err)
		}
		want := strings.TrimSpace("ปรับสต็อกสินค้า (นับสต็อก) จำนวน -2 TAB — " + got.CountNo + " " + note)
		h := f.histories()
		if len(h) != 1 || h[0].Description != want {
			t.Fatalf("note %q: histories %+v, want one row %q", note, h, want)
		}
		if f.count("stock_adjustments", bson.M{"note": got.CountNo + " " + note}) != 1 {
			t.Fatalf("note %q: the Adjustment's own note changed", note)
		}
	}
}
