package ledger

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// ADR-0001: the Stock ledger is the only way a Stock's quantity changes, a
// Stock is created or deleted, a Line's stocks[] or oversoldQty moves, or Sold
// first moves (ADR-0003). This test reads
// every Go file under app/ outside this package and fails on any such write.
// A function may be listed below only with the reason it is still allowed.
var allowedOutsideTheLedger = map[string]string{
	"data/repositories/product_stock.go:createProductStockWithContext": "opening Stock of a Product created in the same transaction: no Line can owe a Product that did not exist",
	"data/repositories/product.go:ClearQuantitySoldFirstById":          "clearing Sold first by hand (ADR-0003)",
}

// setsQuantity finds an update document built before the call, e.g. a
// pipeline {"$set": {"quantity": ...}} held in a variable.
var (
	setsQuantity  = regexp.MustCompile(`"\$(set|inc)"[^"]{0,40}"quantity"`)
	setsSoldFirst = regexp.MustCompile(`"\$(set|inc)"[^"]{0,40}"soldFirst"`)
	setsLineDraw  = regexp.MustCompile(`"\$(set|inc|push)"[^"]{0,40}"(stocks|oversoldQty)"`)
)

var writes = map[string]bool{
	"InsertOne": true, "InsertMany": true, "UpdateOne": true, "UpdateMany": true, "FindOneAndUpdate": true,
	"ReplaceOne": true, "FindOneAndReplace": true, "DeleteOne": true, "DeleteMany": true, "FindOneAndDelete": true, "BulkWrite": true,
}

func TestNothingOutsideTheLedgerChangesStock(t *testing.T) {
	appDir, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	err = filepath.WalkDir(appDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == filepath.Join(appDir, "data", "ledger") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(appDir, path)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || !writes[sel.Sel.Name] {
					return true
				}
				text := string(src[fset.Position(call.Pos()).Offset:fset.Position(call.End()).Offset])
				receiver := string(src[fset.Position(sel.X.Pos()).Offset:fset.Position(sel.X.End()).Offset])
				onStocks := strings.Contains(receiver, "productStockRepo") || strings.Contains(receiver, `"product_stocks"`)
				onProducts := strings.Contains(receiver, "productsRepo") || strings.Contains(receiver, `"products"`)
				// Inserts and deletes always change what Stock exists; an update
				// or bulk write counts when it touches quantity (a BulkWrite's
				// models are built elsewhere in the function).
				updates := strings.Contains(sel.Sel.Name, "Update") || sel.Sel.Name == "BulkWrite"
				body := string(src[fset.Position(fn.Body.Pos()).Offset:fset.Position(fn.Body.End()).Offset])
				touchesQuantity := strings.Contains(text, `"quantity"`) || setsQuantity.MatchString(body)
				changesStock := onStocks && (!updates || touchesQuantity)
				movesSoldFirst := onProducts && (strings.Contains(text, `"soldFirst"`) || setsSoldFirst.MatchString(body))
				onLines := strings.Contains(receiver, "orderItemRepo") || strings.Contains(receiver, `"order_items"`)
				inserts := strings.HasPrefix(sel.Sel.Name, "Insert")
				movesLineDraw := onLines && (inserts || strings.Contains(text, `"oversoldQty"`) || setsLineDraw.MatchString(body))
				if changesStock || movesSoldFirst || movesLineDraw {
					found[filepath.ToSlash(rel)+":"+fn.Name.Name] = true
				}
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var offenders []string
	for site := range found {
		if _, ok := allowedOutsideTheLedger[site]; !ok {
			offenders = append(offenders, site)
		}
	}
	sort.Strings(offenders)
	if len(offenders) > 0 {
		t.Fatalf("Stock or Sold first is written outside the Stock ledger (ADR-0001) by:\n  %s\nMove the change into app/data/ledger.", strings.Join(offenders, "\n  "))
	}
	for site := range allowedOutsideTheLedger {
		if !found[site] {
			t.Errorf("%s no longer writes Stock: remove it from allowedOutsideTheLedger", site)
		}
	}
}
