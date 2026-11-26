package ledger

import (
	"context"
	"errors"
	"strings"
	"testing"

	"pos/app/data/entities"
	"pos/app/domain/constant"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (f *fixture) productDoc() entities.Product {
	f.t.Helper()
	var got entities.Product
	if err := f.pos.Collection("products").FindOne(context.Background(), bson.M{"_id": f.product}).Decode(&got); err != nil {
		f.t.Fatal(err)
	}
	return got
}

func TestDeleteProductZeroesEveryStockWithHistoryAndAnAdjustment(t *testing.T) {
	f := newFixture(t)
	tabs, boxes, empty := f.stock(f.tab, 1, 5), f.stock(f.box, 1, 2), f.stock(f.tab, 2, 0)

	deleted, err := f.ledger.DeleteProduct(context.Background(), f.product.Hex(), "admin")
	if err != nil {
		t.Fatal(err)
	}

	if deleted.DeletedDate == nil || deleted.Id != f.product {
		t.Fatalf("returned product %+v, want it with deletedDate", deleted)
	}
	stored := f.productDoc()
	if stored.DeletedDate == nil || stored.UpdatedBy != "admin" {
		t.Fatalf("stored product %+v, want deletedDate and updatedBy admin", stored)
	}
	if f.quantity(tabs) != 0 || f.quantity(boxes) != 0 || f.quantity(empty) != 0 {
		t.Fatal("every Stock must be zero")
	}
	want := map[string]bool{"ลบสินค้า: ปรับสต็อกเป็น 0 จำนวน -5 TAB": false, "ลบสินค้า: ปรับสต็อกเป็น 0 จำนวน -2 BOX": false}
	histories := f.histories()
	for _, h := range histories {
		if _, ok := want[h.Description]; ok && h.Balance == 0 && h.CreatedBy == "admin" && h.Type == constant.HistoryTypeStockAdjustment {
			want[h.Description] = true
		}
	}
	for d, seen := range want {
		if !seen || len(histories) != 2 {
			t.Fatalf("want one history row %q (balance 0), got %+v", d, histories)
		}
	}
	for st, delta := range map[primitive.ObjectID]int{tabs: -5, boxes: -2} {
		if f.count("stock_adjustments", bson.M{"stockId": st, "delta": delta, "after": 0, "reason": constant.AdjustmentReasonOther, "note": "ลบสินค้า"}) != 1 {
			t.Fatalf("Stock %s: want one Adjustment of %d naming the delete", st.Hex(), delta)
		}
	}
	if f.count("stock_adjustments", bson.M{"stockId": empty}) != 0 {
		t.Fatal("an empty Stock needs no Adjustment")
	}
}

func TestDeleteProductBringsAnOversoldStockUpToZero(t *testing.T) {
	f := newFixture(t)
	oversold := f.stock(f.tab, 1, -3)

	if _, err := f.ledger.DeleteProduct(context.Background(), f.product.Hex(), "admin"); err != nil {
		t.Fatal(err)
	}

	histories := f.histories()
	if f.quantity(oversold) != 0 || len(histories) != 1 || histories[0].Description != "ลบสินค้า: ปรับสต็อกเป็น 0 จำนวน 3 TAB" {
		t.Fatalf("Stock %d, histories %+v: want 0 and one row of +3", f.quantity(oversold), histories)
	}
}

func TestDeleteProductTwiceIsNotFound(t *testing.T) {
	f := newFixture(t)
	if _, err := f.ledger.DeleteProduct(context.Background(), f.product.Hex(), "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := f.ledger.DeleteProduct(context.Background(), f.product.Hex(), "admin"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete: want ErrNotFound, got %v", err)
	}
	if _, err := f.ledger.DeleteProduct(context.Background(), primitive.NewObjectID().Hex(), "admin"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown product: want ErrNotFound, got %v", err)
	}
	var rejected *Rejected
	if _, err := f.ledger.DeleteProduct(context.Background(), "nope", "admin"); !errors.As(err, &rejected) {
		t.Fatalf("bad id: want Rejected, got %v", err)
	}
}

func TestDeleteProductChangesNothingWhenAWriteFails(t *testing.T) {
	f := newFixture(t)
	tabs, boxes := f.stock(f.tab, 1, 5), f.stock(f.box, 1, 2)
	f.failInserts("product_histories")

	_, err := f.ledger.DeleteProduct(context.Background(), f.product.Hex(), "admin")

	if err == nil || !strings.Contains(err.Error(), "validation") {
		t.Fatalf("want the history insert to fail, got %v", err)
	}
	if f.quantity(tabs) != 5 || f.quantity(boxes) != 2 || f.productDoc().DeletedDate != nil || f.count("stock_adjustments", bson.M{}) != 0 {
		t.Fatal("a failed delete must leave Stock, Adjustments and the Product as they were")
	}
}

func TestDeleteProductRefusesAProductInAPendingTransfer(t *testing.T) {
	settle := map[string]func(l *Ledger, id string) (*entities.StockTransfer, error){
		"approved": func(l *Ledger, id string) (*entities.StockTransfer, error) {
			return l.ApproveTransfer(context.Background(), id, "manager")
		},
		"rejected": func(l *Ledger, id string) (*entities.StockTransfer, error) {
			return l.RejectTransfer(context.Background(), id, "manager")
		},
	}
	for name, settleTransfer := range settle {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			source := f.stock(f.tab, 1, 5)
			tr, err := f.ledger.RequestTransfer(context.Background(), f.transfer(primitive.NewObjectID(), source, 3))
			if err != nil {
				t.Fatal(err)
			}
			historiesBefore := len(f.histories())

			_, err = f.ledger.DeleteProduct(context.Background(), f.product.Hex(), "admin")

			var rejected *Rejected
			if !errors.As(err, &rejected) || !strings.Contains(err.Error(), "สินค้านี้อยู่ในใบโอนที่รออนุมัติ ลบไม่ได้") {
				t.Fatalf("delete during a pending Transfer: want the ledger's rejection, got %v", err)
			}
			if f.productDoc().DeletedDate != nil || f.quantity(source) != 2 || len(f.histories()) != historiesBefore || f.count("stock_adjustments", bson.M{}) != 0 {
				t.Fatal("a refused delete must change nothing")
			}
			if _, err := settleTransfer(f.ledger, tr.Id.Hex()); err != nil {
				t.Fatal(err)
			}
			if _, err := f.ledger.DeleteProduct(context.Background(), f.product.Hex(), "admin"); err != nil {
				t.Fatalf("delete after the Transfer is %s: %v", name, err)
			}
		})
	}
}
