package sale

import (
	"encoding/json"
	"os"
	"sort"
	"testing"

	"pos/app/data/entities"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ringCases is testdata/ring_cases.json: which price type a Line charges,
// at what unit price, and which Stock it sells from first. The till's
// preview reads a copy of the same file.
type ringCases struct {
	Note  string `json:"note"`
	Cases []struct {
		Name      string `json:"name"`
		PriceType string `json:"priceType"`
		StockId   string `json:"stockId"`
		Stocks    []struct {
			Id       string  `json:"id"`
			Sequence int     `json:"sequence"`
			Quantity int     `json:"quantity"`
			Price    float64 `json:"price"`
		} `json:"stocks"`
		Prices []struct {
			CustomerType string  `json:"customerType"`
			Price        float64 `json:"price"`
		} `json:"prices"`
		Want struct {
			PriceType  string  `json:"priceType"`
			UnitPrice  float64 `json:"unitPrice"`
			FirstStock string  `json:"firstStock"`
		} `json:"want"`
	} `json:"cases"`
}

const ringCasesPath = "testdata/ring_cases.json"

// ids maps a case's short Stock names to ObjectIDs and back.
func ids(names []string) (map[string]primitive.ObjectID, map[primitive.ObjectID]string) {
	to, back := map[string]primitive.ObjectID{}, map[primitive.ObjectID]string{}
	for _, n := range names {
		id := primitive.NewObjectID()
		to[n], back[id] = id, n
	}
	return to, back
}

func TestRingCases(t *testing.T) {
	raw, err := os.ReadFile(ringCasesPath)
	if err != nil {
		t.Fatal(err)
	}
	var cases ringCases
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatal(err)
	}
	for i, c := range cases.Cases {
		names := []string{}
		for _, s := range c.Stocks {
			names = append(names, s.Id)
		}
		to, back := ids(names)
		var stocks []entities.ProductStock
		for _, s := range c.Stocks {
			stocks = append(stocks, entities.ProductStock{Id: to[s.Id], Sequence: s.Sequence, Quantity: s.Quantity, Price: s.Price})
		}
		var prices []entities.ProductPrice
		for _, p := range c.Prices {
			prices = append(prices, entities.ProductPrice{CustomerType: p.CustomerType, Price: p.Price})
		}
		chosen := ""
		if c.StockId != "" {
			chosen = to[c.StockId].Hex()
		}
		line := Line{PriceType: c.PriceType, StockId: chosen, Quantity: 1}
		// As Ring orders them: by sequence, keeping the given order on a tie.
		sort.SliceStable(stocks, func(a, b int) bool { return stocks[a].Sequence < stocks[b].Sequence })
		first := firstStock(line, stocks)
		gotType, gotPrice := price(line, prices, first)
		gotFirst := ""
		if first != nil {
			gotFirst = back[first.Id]
		}
		if *update {
			cases.Cases[i].Want.PriceType, cases.Cases[i].Want.UnitPrice, cases.Cases[i].Want.FirstStock = gotType, gotPrice, gotFirst
			continue
		}
		if gotType != c.Want.PriceType || gotPrice != c.Want.UnitPrice || gotFirst != c.Want.FirstStock {
			t.Errorf("%s: got %q %v from %q, want %q %v from %q", c.Name, gotType, gotPrice, gotFirst,
				c.Want.PriceType, c.Want.UnitPrice, c.Want.FirstStock)
		}
	}
	if *update {
		out, _ := json.MarshalIndent(cases, "", "  ")
		if err := os.WriteFile(ringCasesPath, append(out, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
