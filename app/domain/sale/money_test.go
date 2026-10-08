package sale

import (
	"encoding/json"
	"flag"
	"os"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/money_cases.json from this package")

// moneyCases is testdata/money_cases.json. The till's tests read a copy of
// the same file, so a rule changed on one side fails the other.
type moneyCases struct {
	Note  string `json:"note"`
	Lines []struct {
		Name      string  `json:"name"`
		UnitPrice float64 `json:"unitPrice"`
		Quantity  int     `json:"quantity"`
		Discount  float64 `json:"discount"`
		Want      Charge  `json:"want"`
	} `json:"lines"`
	Sales []struct {
		Name  string `json:"name"`
		Lines []struct {
			UnitPrice float64 `json:"unitPrice"`
			Quantity  int     `json:"quantity"`
			Discount  float64 `json:"discount"`
		} `json:"lines"`
		WantTotal float64 `json:"wantTotal"`
	} `json:"sales"`
	Tenders []struct {
		Name       string  `json:"name"`
		Total      float64 `json:"total"`
		Tendered   float64 `json:"tendered"`
		WantCovers bool    `json:"wantCovers"`
		WantChange float64 `json:"wantChange"`
	} `json:"tenders"`
}

const moneyCasesPath = "testdata/money_cases.json"

func TestMoneyCases(t *testing.T) {
	raw, err := os.ReadFile(moneyCasesPath)
	if err != nil {
		t.Fatal(err)
	}
	var cases moneyCases
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}

	for i, c := range cases.Lines {
		got := ChargeFor(c.UnitPrice, c.Quantity, c.Discount)
		if *update {
			cases.Lines[i].Want = got
		} else if got != c.Want {
			t.Errorf("line %q: got %+v, want %+v", c.Name, got, c.Want)
		}
	}
	for i, c := range cases.Sales {
		var totals Totals
		for _, l := range c.Lines {
			charge := ChargeFor(l.UnitPrice, l.Quantity, l.Discount)
			totals.Add(charge.Paid, 0, charge.Discount*float64(l.Quantity))
		}
		got, _, _ := totals.Rounded()
		if *update {
			cases.Sales[i].WantTotal = got
		} else if got != c.WantTotal {
			t.Errorf("sale %q: total %v, want %v", c.Name, got, c.WantTotal)
		}
	}
	for i, c := range cases.Tenders {
		change, covers := Tender(c.Tendered, c.Total)
		if *update {
			cases.Tenders[i].WantCovers, cases.Tenders[i].WantChange = covers, change
		} else if covers != c.WantCovers || change != c.WantChange {
			t.Errorf("tender %q: got covers=%v change=%v, want covers=%v change=%v",
				c.Name, covers, change, c.WantCovers, c.WantChange)
		}
	}

	if *update {
		out, err := json.MarshalIndent(cases, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(moneyCasesPath, append(out, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
