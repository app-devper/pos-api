package promotion

import (
	"errors"
	"testing"
)

func TestTermsMustDescribeARealDiscount(t *testing.T) {
	cases := []struct {
		name  string
		terms Terms
		want  error
	}{
		{"a percentage", Terms{Type: Percentage, Value: 10}, nil},
		{"a fixed amount", Terms{Type: Fixed, Value: 50}, nil},
		{"a type written in lower case is read as the same type", Terms{Type: "percentage", Value: 10}, nil},
		{"an unknown type", Terms{Type: "BOGO", Value: 1}, ErrUnknownType},
		{"no value", Terms{Type: Fixed, Value: 0}, ErrValue},
		{"a negative value", Terms{Type: Fixed, Value: -5}, ErrValue},
		{"more than 100 percent", Terms{Type: Percentage, Value: 150}, ErrValue},
		{"a negative cap", Terms{Type: Percentage, Value: 10, MaxDiscount: -1}, ErrValue},
		{"a negative minimum", Terms{Type: Fixed, Value: 5, MinPurchase: -1}, ErrValue},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.terms.Validate(); !errors.Is(err, c.want) {
				t.Fatalf("got %v, want %v", err, c.want)
			}
		})
	}
}

func TestDiscount(t *testing.T) {
	cases := []struct {
		name     string
		terms    Terms
		total    float64
		products []string
		want     float64
		err      error
	}{
		{"a percentage of the total", Terms{Type: Percentage, Value: 10}, 333.33, nil, 33.33, nil},
		{"capped by the most it may give", Terms{Type: Percentage, Value: 50, MaxDiscount: 100}, 1000, nil, 100, nil},
		{"a fixed amount", Terms{Type: Fixed, Value: 50}, 200, nil, 50, nil},
		{"never more than the total", Terms{Type: Fixed, Value: 500}, 200, nil, 200, nil},
		{"below the minimum purchase", Terms{Type: Fixed, Value: 50, MinPurchase: 300}, 200, nil, 0, ErrBelowMinimum},
		{"limited to products the Sale does not hold", Terms{Type: Fixed, Value: 50, ProductIds: []string{"a"}}, 200, []string{"b"}, 0, ErrNoMatchingProduct},
		{"limited to a product the Sale holds", Terms{Type: Fixed, Value: 50, ProductIds: []string{"a"}}, 200, []string{"b", "a"}, 50, nil},
		{"a lower-case type still discounts", Terms{Type: "percentage", Value: 10}, 100, nil, 10, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := c.terms.Discount(c.total, c.products)
			if !errors.Is(err, c.err) || got != c.want {
				t.Fatalf("got %v, %v; want %v, %v", got, err, c.want, c.err)
			}
		})
	}
}
