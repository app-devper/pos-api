package repositories

import (
	"testing"

	"pos/app/domain/constant"
)

func TestGetSequenceDefinition(t *testing.T) {
	testCases := []struct {
		name   string
		field  string
		prefix string
		kind   string
	}{
		{name: "order daily", field: constant.ORDER, prefix: "OD_", kind: constant.DAILY},
		{name: "receive daily", field: constant.RECEIVE, prefix: "RC_", kind: constant.DAILY},
		{name: "member yearly", field: constant.MEMBER, prefix: "MB_", kind: constant.YEARLY},
		{name: "product none", field: constant.PRODUCT, prefix: "PD_", kind: constant.NONE},
		{name: "default none", field: "UNKNOWN", prefix: "", kind: constant.NONE},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			definition := getSequenceDefinition(tc.field)
			if definition.Prefix != tc.prefix {
				t.Fatalf("expected prefix %s, got %s", tc.prefix, definition.Prefix)
			}
			if definition.Type != tc.kind {
				t.Fatalf("expected type %s, got %s", tc.kind, definition.Type)
			}
			if definition.Format != 4 {
				t.Fatalf("expected format 4, got %d", definition.Format)
			}
		})
	}
}

func TestBuildNextSequenceValueExpression(t *testing.T) {
	if expr := buildNextSequenceValueExpression(constant.NONE, "20250101"); expr == nil {
		t.Fatal("expected expression for none sequence")
	}
	if expr := buildNextSequenceValueExpression(constant.DAILY, "20250101"); expr == nil {
		t.Fatal("expected expression for daily sequence")
	}
	if expr := buildNextSequenceValueExpression(constant.MONTHLY, "20250101"); expr == nil {
		t.Fatal("expected expression for monthly sequence")
	}
	if expr := buildNextSequenceValueExpression(constant.YEARLY, "20250101"); expr == nil {
		t.Fatal("expected expression for yearly sequence")
	}
}

func TestBuildNextSequenceDateExpression(t *testing.T) {
	if expr := buildNextSequenceDateExpression(constant.NONE, "20250101"); expr == nil {
		t.Fatal("expected date expression for none sequence")
	}
	if expr := buildNextSequenceDateExpression(constant.DAILY, "20250101"); expr == nil {
		t.Fatal("expected date expression for daily sequence")
	}
	if expr := buildNextSequenceDateExpression(constant.MONTHLY, "20250101"); expr == nil {
		t.Fatal("expected date expression for monthly sequence")
	}
	if expr := buildNextSequenceDateExpression(constant.YEARLY, "20250101"); expr == nil {
		t.Fatal("expected date expression for yearly sequence")
	}
}
